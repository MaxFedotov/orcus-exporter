package client

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/go-sql-driver/mysql"
	ini "gopkg.in/ini.v1"
)

// XtradbClient allows you to get Xtradb cluster metrics.
type XtradbClient struct {
	dsn string
}

// XtradbMetrics represents Xtradb cluster metrics.
type XtradbMetrics struct {
	ClusterSize   int
	NodeState     int
	ClusterStatus int
}

// NewXtradbClient creates an XtradbClient.
func NewXtradbClient(myCnf string, sslVerify bool) (*XtradbClient, error) {
	dsn, err := parseMycnf(myCnf, sslVerify)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse my.cnf for Xtradb cluster client: %v", err)
	}

	client := &XtradbClient{
		dsn: dsn,
	}

	if _, err := client.GetMetrics(); err != nil {
		return nil, fmt.Errorf("Failed to create Xtradb cluster client: %v", err)
	}

	return client, nil
}

func parseMycnf(myCnf string, sslVerify bool) (string, error) {
	var dsn string
	opts := ini.LoadOptions{
		// MySQL ini file can have boolean keys.
		AllowBooleanKeys: true,
	}
	cfg, err := ini.LoadSources(opts, myCnf)
	if err != nil {
		return dsn, fmt.Errorf("failed reading ini file: %s", err)
	}
	user := cfg.Section("client").Key("user").String()
	password := cfg.Section("client").Key("password").String()
	if (user == "") || (password == "") {
		return dsn, fmt.Errorf("no user or password specified under [client] in %s", myCnf)
	}
	host := cfg.Section("client").Key("host").MustString("localhost")
	port := cfg.Section("client").Key("port").MustUint(3306)
	socket := cfg.Section("client").Key("socket").String()
	if socket != "" {
		dsn = fmt.Sprintf("%s:%s@unix(%s)/", user, password, socket)
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/", user, password, host, port)
	}
	sslCA := cfg.Section("client").Key("ssl-ca").String()
	sslCert := cfg.Section("client").Key("ssl-cert").String()
	sslKey := cfg.Section("client").Key("ssl-key").String()
	if sslCA != "" {
		if tlsErr := customizeTLS(sslCA, sslCert, sslKey, sslVerify); tlsErr != nil {
			tlsErr = fmt.Errorf("failed to register a custom TLS configuration for mysql dsn: %s", tlsErr)
			return dsn, tlsErr
		}
		dsn = fmt.Sprintf("%s?tls=custom", dsn)
	}

	return dsn, nil
}

func customizeTLS(sslCA string, sslCert string, sslKey string, sslVerify bool) error {
	var tlsCfg tls.Config
	caBundle := x509.NewCertPool()
	pemCA, err := ioutil.ReadFile(sslCA)
	if err != nil {
		return err
	}
	if ok := caBundle.AppendCertsFromPEM(pemCA); ok {
		tlsCfg.RootCAs = caBundle
	} else {
		return fmt.Errorf("failed parse pem-encoded CA certificates from %s", sslCA)
	}
	if sslCert != "" && sslKey != "" {
		certPairs := make([]tls.Certificate, 0, 1)
		keypair, err := tls.LoadX509KeyPair(sslCert, sslKey)
		if err != nil {
			return fmt.Errorf("failed to parse pem-encoded SSL cert %s or SSL key %s: %s",
				sslCert, sslKey, err)
		}
		certPairs = append(certPairs, keypair)
		tlsCfg.Certificates = certPairs
		tlsCfg.InsecureSkipVerify = !sslVerify
	}
	mysql.RegisterTLSConfig("custom", &tlsCfg)
	return nil
}

// GetMetrics fetches Xtradb cluster metrics.
func (client *XtradbClient) GetMetrics() (*XtradbMetrics, error) {
	var clusterSize, nodeState, clusterStatus int
	var clusterStatusDesc string
	var variableName string
	var metrics XtradbMetrics
	db, err := sql.Open("mysql", client.dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection to database: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(1 * time.Minute)
	err = db.QueryRow("SHOW STATUS LIKE 'wsrep_cluster_size';").Scan(&variableName, &clusterSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get data from database: %v", err)
	}
	err = db.QueryRow("SHOW STATUS LIKE 'wsrep_local_state';").Scan(&variableName, &nodeState)
	if err != nil {
		return nil, fmt.Errorf("failed to get data from database: %v", err)
	}
	err = db.QueryRow("SHOW STATUS LIKE 'wsrep_cluster_status';").Scan(&variableName, &clusterStatusDesc)
	if err != nil {
		return nil, fmt.Errorf("failed to get data from database: %v", err)
	}
	switch clusterStatusDesc {
	case "Primary":
		clusterStatus = 1
	default:
		clusterStatus = 0
	}
	metrics.ClusterSize = clusterSize
	metrics.NodeState = nodeState
	metrics.ClusterStatus = clusterStatus
	return &metrics, nil
}
