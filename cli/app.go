package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/btcsuite/btcutil"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/notifier"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/server"
	"github.com/qtumproject/janus/pkg/transformer"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("janus", "Qtum adapter to Ethereum JSON RPC")

	accountsFile = app.Flag("accounts", "account private keys (in WIF) returned by eth_accounts").Envar("ACCOUNTS").File()

	qtumRPC     = app.Flag("qtum-rpc", "URL of qtum RPC service").Envar("QTUM_RPC").Default("").String()
	qtumNetwork = app.Flag("qtum-network", "").Envar("QTUM_NETWORK").Default("regtest").String()
	bind        = app.Flag("bind", "network interface to bind to (e.g. 0.0.0.0) ").Default("localhost").String()
	port        = app.Flag("port", "port to serve proxy").Default("23889").Int()
	httpsKey    = app.Flag("https-key", "https keyfile").Default("").String()
	httpsCert   = app.Flag("https-cert", "https certificate").Default("").String()
	logFile     = app.Flag("log-file", "write logs to a file").Envar("LOG_FILE").Default("").String()

	devMode        = app.Flag("dev", "[Insecure] Developer mode").Envar("DEV").Default("false").Bool()
	singleThreaded = app.Flag("singleThreaded", "[Non-production] Process RPC requests in a single thread").Envar("SINGLE_THREADED").Default("false").Bool()

	ignoreUnknownTransactions = app.Flag("ignoreTransactions", "[Development] Ignore transactions inside blocks we can't fetch and return responses instead of failing").Default("false").Bool()
	disableSnipping           = app.Flag("disableSnipping", "[Development] Disable ...snip... in logs").Default("false").Bool()
	hideQtumdLogs             = app.Flag("hideQtumdLogs", "[Development] Hide QTUMD debug logs").Default("false").Bool()

	generateToAddressTo = app.Flag("generateToAddressTo", "[regtest only] configure address to mine blocks to when mining new transactions in blocks").Envar("GENERATE_TO_ADDRESS").Default("").String()
)

func loadAccounts(r io.Reader, l log.Logger) qtum.Accounts {
	var accounts qtum.Accounts

	if accountsFile != nil {
		s := bufio.NewScanner(*accountsFile)
		for s.Scan() {
			line := s.Text()

			wif, err := btcutil.DecodeWIF(line)
			if err != nil {
				level.Error(l).Log("msg", "Failed to parse account", "err", err.Error())
				continue
			}

			accounts = append(accounts, wif)
		}
	}

	if len(accounts) > 0 {
		level.Info(l).Log("msg", fmt.Sprintf("Loaded %d accounts", len(accounts)))
	} else {
		level.Warn(l).Log("msg", "No accounts loaded from account file")
	}

	return accounts
}

func action(pc *kingpin.ParseContext) error {
	addr := fmt.Sprintf("%s:%d", *bind, *port)
	writers := []io.Writer{os.Stdout}

	if logFile != nil && (*logFile) != "" {
		_, err := os.Stat(*logFile)
		if os.IsNotExist(err) {
			newLogFile, err := os.Create(*logFile)
			if err != nil {
				return errors.Wrapf(err, "Failed to create log file %s", *logFile)
			} else {
				writers = append(writers, newLogFile)
			}
		} else {
			existingLogFile, err := os.Open(*logFile)
			if err != nil {
				return errors.Wrapf(err, "Failed to open log file %s", *logFile)
			} else {
				writers = append(writers, existingLogFile)
			}
		}
	}

	logWriter := io.MultiWriter(writers...)
	logger := log.NewLogfmtLogger(logWriter)

	if !*devMode {
		logger = level.NewFilter(logger, level.AllowWarn())
	}

	var accounts qtum.Accounts
	if *accountsFile != nil {
		accounts = loadAccounts(*accountsFile, logger)
		(*accountsFile).Close()
	}

	isMain := *qtumNetwork == qtum.ChainMain

	qtumJSONRPC, err := qtum.NewClient(
		isMain,
		*qtumRPC,
		qtum.SetDebug(*devMode),
		qtum.SetLogWriter(logWriter),
		qtum.SetLogger(logger),
		qtum.SetAccounts(accounts),
		qtum.SetGenerateToAddress(*generateToAddressTo),
		qtum.SetIgnoreUnknownTransactions(*ignoreUnknownTransactions),
		qtum.SetDisableSnippingQtumRpcOutput(*disableSnipping),
		qtum.SetHideQtumdLogs(*hideQtumdLogs),
	)
	if err != nil {
		return errors.Wrap(err, "jsonrpc#New")
	}

	qtumClient, err := qtum.New(qtumJSONRPC, *qtumNetwork)
	if err != nil {
		return errors.Wrap(err, "qtum#New")
	}

	agent := notifier.NewAgent(context.Background(), qtumClient, nil)
	proxies := transformer.DefaultProxies(qtumClient, agent)
	t, err := transformer.New(
		qtumClient,
		proxies,
		transformer.SetDebug(*devMode),
		transformer.SetLogger(logger),
	)
	if err != nil {
		return errors.Wrap(err, "transformer#New")
	}
	agent.SetTransformer(t)

	httpsKeyFile := getEmptyStringIfFileDoesntExist(*httpsKey, logger)
	httpsCertFile := getEmptyStringIfFileDoesntExist(*httpsCert, logger)

	s, err := server.New(
		qtumClient,
		t,
		addr,
		server.SetLogWriter(logWriter),
		server.SetLogger(logger),
		server.SetDebug(*devMode),
		server.SetSingleThreaded(*singleThreaded),
		server.SetHttps(httpsKeyFile, httpsCertFile),
	)
	if err != nil {
		return errors.Wrap(err, "server#New")
	}

	return s.Start()
}

func getEmptyStringIfFileDoesntExist(file string, l log.Logger) string {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		l.Log("file does not exist", file)
		return ""
	}
	return file
}

func Run() {
	kingpin.MustParse(app.Parse(os.Args[1:]))
}

func init() {
	app.Action(action)
}
