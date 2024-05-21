package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "nextcloud"
)

var (
	promDescResult = prometheus.NewDesc(
		namespace+"_result_ok",
		"1 if the last scrape is successful.",
		nil, nil)
	promDescSystemInfo = prometheus.NewDesc(
		namespace+"_system_info",
		"Information about Nextcloud installation.",
		[]string{"version", "php_version", "webserver", "database", "database_version"}, nil)
	promDescApps = prometheus.NewDesc(
		namespace+"_num_apps",
		"Number applications installed and with update available.",
		[]string{"status"}, nil)
	promDescUpdate = prometheus.NewDesc(
		namespace+"_update",
		"Update available.",
		nil, nil)
	promDescNumUsers = prometheus.NewDesc(
		namespace+"_num_users_total",
		"Number of users on the instance.",
		[]string{"status"}, nil)
	promDescNumFiles = prometheus.NewDesc(
		namespace+"_num_files_total",
		"Number of files served by the instance.",
		nil, nil)
	promDescNumStorages = prometheus.NewDesc(
		namespace+"_num_storages_total",
		"Number of storages served by the instance.",
		[]string{"type"}, nil)
	promDescFreeSpace = prometheus.NewDesc(
		namespace+"_free_space_bytes",
		"Free space on the instance in bytes.",
		nil, nil)
	promDescShares = prometheus.NewDesc(
		namespace+"_shares_total",
		"Number of shares by type.",
		[]string{"type"}, nil)
	promDescFedShares = prometheus.NewDesc(
		namespace+"_fed_shares_total",
		"Number of federated shares by direction.",
		[]string{"direction"}, nil)
	promDescPhpMaxExecutionTime = prometheus.NewDesc(
		namespace+"_max_execution_time_seconds",
		"PHP max execution time in seconds.",
		nil, nil)
	promDescPhpMemoryLimit = prometheus.NewDesc(
		namespace+"_php_memory_limit_bytes",
		"PHP memory limit in bytes.",
		nil, nil)
	promDescPhpMaxUploadSize = prometheus.NewDesc(
		namespace+"_php_upload_max_size_bytes",
		"PHP maximum upload size in bytes.",
		nil, nil)
	promDescDatabaseSize = prometheus.NewDesc(
		namespace+"_db_size_bytes",
		"Database size in bytes.",
		nil, nil)
)

type Collector struct {
	client         *http.Client
	infoURL        string
	user, password string

	promRequests   *prometheus.CounterVec
	promRequestsOk prometheus.Counter
	promRequestsKo prometheus.Counter
}

func NewCollector(infoURL url.URL, client *http.Client) Collector {
	urlUser := infoURL.User
	infoURL.User = nil
	user, password := "", ""
	if urlUser != nil {
		user = urlUser.Username()
		password, _ = urlUser.Password()
	}
	log.Printf("serverinfo URL (no user/password): %s\n", infoURL.String())
	promRequests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "requests",
		Help:      "Counts the number of requests to the exporter",
	}, []string{"status"})

	return Collector{
		client:         client,
		infoURL:        infoURL.String(),
		user:           user,
		password:       password,
		promRequests:   promRequests,
		promRequestsOk: promRequests.WithLabelValues("ok"),
		promRequestsKo: promRequests.WithLabelValues("ko"),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- promDescSystemInfo
	ch <- promDescApps
	ch <- promDescUpdate
	ch <- promDescNumUsers
	ch <- promDescNumFiles
	ch <- promDescNumStorages
	ch <- promDescShares
	ch <- promDescFedShares
	ch <- promDescPhpMaxExecutionTime
	ch <- promDescPhpMemoryLimit
	ch <- promDescPhpMaxUploadSize
	ch <- promDescDatabaseSize
	c.promRequests.Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	resultOk := 0.
	if nextcloud, err := c.retrieveNextcloudData(); err == nil {
		c.collectMetrics(nextcloud, ch)
		resultOk = 1.
		c.promRequestsOk.Inc()
	} else {
		log.Printf("Error: %s\n", err)
		c.promRequestsKo.Inc()
	}
	ch <- prometheus.MustNewConstMetric(promDescResult, prometheus.GaugeValue, resultOk)
	c.promRequestsOk.Collect(ch)
	c.promRequestsKo.Collect(ch)
}

func (c *Collector) retrieveNextcloudData() (*NextCloudRoot, error) {
	request, err := http.NewRequest(http.MethodGet, c.infoURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Accept", "application/json")
	if c.user != "" || c.password != "" {
		request.SetBasicAuth(c.user, c.password)
	}
	res, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Got HTTP %d %s", res.StatusCode, res.Status)
	}

	jsonResult := &NextCloudRoot{}
	return jsonResult, json.NewDecoder(res.Body).Decode(jsonResult)
}
func (c *Collector) collectMetrics(nextcloud *NextCloudRoot, ch chan<- prometheus.Metric) {
	data := &nextcloud.Ocs.Data
	ch <- prometheus.MustNewConstMetric(promDescSystemInfo, prometheus.GaugeValue,
		float64(1),
		data.Nextcloud.System.Version,
		data.Server.PHP.Version,
		data.Server.Webserver,
		data.Server.Database.Type,
		data.Server.Database.Version)

	if data.Nextcloud.System.Apps != nil {
		ch <- prometheus.MustNewConstMetric(promDescApps, prometheus.GaugeValue,
			float64(data.Nextcloud.System.Apps.NumInstalled),
			"installed")
		ch <- prometheus.MustNewConstMetric(promDescApps, prometheus.GaugeValue,
			float64(data.Nextcloud.System.Apps.NumUpdatesAvailable),
			"updates_available")
	}
	if data.Nextcloud.System.Update != nil {
		ch <- prometheus.MustNewConstMetric(promDescUpdate, prometheus.GaugeValue,
			boolToFloat(data.Nextcloud.System.Update.Available))
	}
	ch <- prometheus.MustNewConstMetric(promDescNumUsers, prometheus.GaugeValue,
		float64(data.ActiveUsers.Last5Minutes),
		"active")
	ch <- prometheus.MustNewConstMetric(promDescNumUsers, prometheus.GaugeValue,
		float64(data.Nextcloud.Storage.NumUsers),
		"registered")
	ch <- prometheus.MustNewConstMetric(promDescNumFiles, prometheus.GaugeValue,
		float64(data.Nextcloud.Storage.NumFiles))
	ch <- prometheus.MustNewConstMetric(promDescNumStorages, prometheus.GaugeValue,
		float64(data.Nextcloud.Storage.NumStoragesHome),
		"home")
	ch <- prometheus.MustNewConstMetric(promDescNumStorages, prometheus.GaugeValue,
		float64(data.Nextcloud.Storage.NumStoragesLocal),
		"local")
	ch <- prometheus.MustNewConstMetric(promDescNumStorages, prometheus.GaugeValue,
		float64(data.Nextcloud.Storage.NumStoragesOther),
		"other")
	ch <- prometheus.MustNewConstMetric(promDescFreeSpace, prometheus.GaugeValue,
		float64(data.Nextcloud.System.FreeSpace))
	ch <- prometheus.MustNewConstMetric(promDescShares, prometheus.GaugeValue,
		float64(data.Nextcloud.Shares.NumSharesUser),
		"user")
	ch <- prometheus.MustNewConstMetric(promDescShares, prometheus.GaugeValue,
		float64(data.Nextcloud.Shares.NumSharesGroups),
		"groups")
	ch <- prometheus.MustNewConstMetric(promDescShares, prometheus.GaugeValue,
		float64(data.Nextcloud.Shares.NumSharesMail),
		"mail")
	ch <- prometheus.MustNewConstMetric(promDescShares, prometheus.GaugeValue,
		float64(data.Nextcloud.Shares.NumSharesRoom),
		"room")
	ch <- prometheus.MustNewConstMetric(promDescShares, prometheus.GaugeValue,
		float64(data.Nextcloud.Shares.NumSharesLink-data.Nextcloud.Shares.NumSharesLinkNoPassword),
		"link_password")
	ch <- prometheus.MustNewConstMetric(promDescShares, prometheus.GaugeValue,
		float64(data.Nextcloud.Shares.NumSharesLinkNoPassword),
		"link_nopassword")
	ch <- prometheus.MustNewConstMetric(promDescFedShares, prometheus.GaugeValue,
		float64(data.Nextcloud.Shares.NumFedSharesSent),
		"sent")
	ch <- prometheus.MustNewConstMetric(promDescFedShares, prometheus.GaugeValue,
		float64(data.Nextcloud.Shares.NumFedSharesReceived),
		"received")
	ch <- prometheus.MustNewConstMetric(promDescPhpMaxExecutionTime, prometheus.GaugeValue,
		float64(data.Server.PHP.MaxExecutionTime))
	ch <- prometheus.MustNewConstMetric(promDescPhpMemoryLimit, prometheus.GaugeValue,
		float64(data.Server.PHP.MemoryLimit))
	ch <- prometheus.MustNewConstMetric(promDescPhpMaxUploadSize, prometheus.GaugeValue,
		float64(data.Server.PHP.UploadMaxFileSize))
	ch <- prometheus.MustNewConstMetric(promDescDatabaseSize, prometheus.GaugeValue,
		float64(data.Server.Database.Size))
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
