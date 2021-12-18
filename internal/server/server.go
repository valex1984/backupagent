package server

import (
	"backupagent/config"
	"backupagent/internal/client"
	"backupagent/internal/utils"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	lxd "github.com/lxc/lxd/client"
)

type Server struct {
	HttpSrv *http.Server
	s3      *client.S3client
	lxd     *lxd.InstanceServer
	cfg     *config.Config
}

type lxcDataReq struct {
	Name       string
	BackupName string
}

func NewServer(cfg *config.Config) (*Server, error) {

	s3c, err := client.NewS3client(cfg)
	if err != nil {
		return nil, err
	}

	is, err := lxd.ConnectLXDUnix("/var/snap/lxd/common/lxd/unix.socket", nil)
	if err != nil {
		return nil, err
	}

	s := Server{nil, s3c, &is, cfg}
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/backup", s.BasicAuth(s.backupLxcHandler)).Methods("POST")
	router.HandleFunc("/restore", s.BasicAuth(s.restoreLxcHandler)).Methods("POST")
	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	s.HttpSrv = &http.Server{
		Addr:         ":" + cfg.Htts.Port,
		Handler:      router,
		TLSConfig:    tlsConfig,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	return &s, nil
}

func (s *Server) Run() error {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	go func() {
		s.HttpSrv.ListenAndServeTLS(s.cfg.Htts.TlsCertFiel, s.cfg.Htts.TlsKeyFile)
	}()

	<-ctx.Done()
	return nil

}

func (s *Server) BasicAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		u, p, ok := r.BasicAuth()
		if !ok || u != s.cfg.Htts.Username || p != s.cfg.Htts.Password {
			w.WriteHeader(401)
			return
		}
		handler(w, r)
	}
}

func (s *Server) backupLxcHandler(w http.ResponseWriter, r *http.Request) {

	var br lxcDataReq
	err := json.NewDecoder(r.Body).Decode(&br)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := s.backup(br.Name, br.BackupName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(ret)
}

func (s *Server) restoreLxcHandler(w http.ResponseWriter, r *http.Request) {

	var br lxcDataReq
	err := json.NewDecoder(r.Body).Decode(&br)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := s.restore(br.Name, br.BackupName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(ret)
}

func (s *Server) backup(containerName string, name string) (int, error) {

	c, req, err := utils.PrepareLxcBackupRequest(containerName, name, s.lxd)
	if err != nil {
		return 500, err
	}
	response, err := c.Do(req)
	if err != nil {
		return 500, err
	}
	if response.StatusCode >= 300 {
		return response.StatusCode, nil
	}
	defer response.Body.Close()

	var r io.Reader = response.Body
	err = s.s3.Upload(name, &r)
	if err != nil {
		return 500, err
	}
	return 200, nil
}

func (s *Server) restore(containerName string, name string) (int, error) {

	r, err := s.s3.Download(name)
	if err != nil {
		return 500, err
	}
	defer r.Body.Close()

	lxdi := *s.lxd
	_, err = lxdi.CreateContainerFromBackup(
		lxd.ContainerBackupArgs{BackupFile: r.Body})

	if err != nil {
		return 500, err
	}

	// if response.StatusCode >= 300 {
	// 	return response.StatusCode, nil
	// }
	// log.Println(len(buffer.Bytes()))
	return 200, nil
}
