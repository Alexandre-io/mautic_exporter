package main

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"flag"
	"fmt"
	"os"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

//This is my collector metrics
type mauticCollector struct {
	numLeadsMetric        *prometheus.Desc
	numAnonymousMetric    *prometheus.Desc
	numEmailsMetric       *prometheus.Desc
	numCampaignsMetric    *prometheus.Desc
	numPageHitsMetric     *prometheus.Desc
	numWebhookQueueMetric *prometheus.Desc
	numMessageQueueMetric *prometheus.Desc

	dbHost        string
	dbName        string
	dbUser        string
	dbPass        string
	dbTablePrefix string
}

//This is a constructor for my mauticCollector struct
func newMauticCollector(host string, dbname string, username string, pass string, tablePrefix string) *mauticCollector {
	return &mauticCollector{
		numLeadsMetric: prometheus.NewDesc("mautic_num_leads_metric",
			"Shows the number of leads in Mautic",
			nil, nil,
		),
		numAnonymousMetric: prometheus.NewDesc("mautic_num_anonymous_metric",
			"Shows the number of anonymous in Mautic",
			nil, nil,
		),
		numEmailsMetric: prometheus.NewDesc("mautic_num_emails_metric",
			"Shows the number of emails in Mautic",
			nil, nil,
		),
		numCampaignsMetric: prometheus.NewDesc("mautic_num_campaigns_metric",
			"Shows the number of campaigns in Mautic",
			nil, nil,
		),
		numPageHitsMetric: prometheus.NewDesc("mautic_num_pagehits_metric",
			"Shows the number of page hits in Mautic",
			nil, nil,
		),
		numWebhookQueueMetric: prometheus.NewDesc("mautic_num_webhook_queue_metric",
			"Shows the number of webhook in Mautic's queue",
			nil, nil,
		),
		numMessageQueueMetric: prometheus.NewDesc("mautic_num_message_queue_metric",
			"Shows the number of message in Mautic's queue",
			nil, nil,
		),

		dbHost:        host,
		dbName:        dbname,
		dbUser:        username,
		dbPass:        pass,
		dbTablePrefix: tablePrefix,
	}
}

//Describe method is required for a prometheus.Collector type
func (collector *mauticCollector) Describe(ch chan<- *prometheus.Desc) {

	//We set the metrics
	ch <- collector.numLeadsMetric
	ch <- collector.numAnonymousMetric
	ch <- collector.numEmailsMetric
	ch <- collector.numCampaignsMetric
	ch <- collector.numPageHitsMetric
	ch <- collector.numWebhookQueueMetric
	ch <- collector.numMessageQueueMetric

}

//Collect method is required for a prometheus.Collector type
func (collector *mauticCollector) Collect(ch chan<- prometheus.Metric) {

	//We run DB queries here to retrieve the metrics we care about
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", collector.dbUser, collector.dbPass, collector.dbHost, collector.dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %s ...\n", err)
		os.Exit(1)
	}
	defer db.Close()

	//select count(id) as numLeads from leads where date_identified is not null
	var numLeads float64
	q1 := fmt.Sprintf("select count(id) as numLeads from %sleads where date_identified is not null;", collector.dbTablePrefix)
	err = db.QueryRow(q1).Scan(&numLeads)
	if err != nil {
		log.Fatal(err)
	}

	//select count(id) as numAnonymous from leads where date_identified is null
	var numAnonymous float64
	q2 := fmt.Sprintf("select count(id) as numAnonymous from %sleads where date_identified is null;", collector.dbTablePrefix)
	err = db.QueryRow(q2).Scan(&numAnonymous)
	if err != nil {
		log.Fatal(err)
	}

	//select count(*) as numEmails from emails;
	var numEmails float64
	q3 := fmt.Sprintf("select count(*) as numEmails from %semails;", collector.dbTablePrefix)
	err = db.QueryRow(q3).Scan(&numEmails)
	if err != nil {
		log.Fatal(err)
	}

	//select count(*) as numCampaigns from campaigns;
	var numCampaigns float64
	q4 := fmt.Sprintf("select count(*) as numCampaigns from %scampaigns;", collector.dbTablePrefix)
	err = db.QueryRow(q4).Scan(&numCampaigns)
	if err != nil {
		log.Fatal(err)
	}

	//select count(*) as numPageHits from page_hits;
	var numPageHits float64
	q5 := fmt.Sprintf("select count(*) as numPageHits from %spage_hits;", collector.dbTablePrefix)
	err = db.QueryRow(q5).Scan(&numPageHits)
	if err != nil {
		log.Fatal(err)
	}

	//select count(*) as numWebhookQueue from webhook_queue;
	var numWebhookQueue float64
	q6 := fmt.Sprintf("select count(*) as numWebhookQueue from %swebhook_queue;", collector.dbTablePrefix)
	err = db.QueryRow(q6).Scan(&numWebhookQueue)
	if err != nil {
		log.Fatal(err)
	}

	//select count(*) as numMessageQueue from message_queue;
	var numMessageQueue float64
	q7 := fmt.Sprintf("select count(*) as numMessageQueue from %smessage_queue;", collector.dbTablePrefix)
	err = db.QueryRow(q7).Scan(&numMessageQueue)
	if err != nil {
		log.Fatal(err)
	}

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	ch <- prometheus.MustNewConstMetric(collector.numLeadsMetric, prometheus.CounterValue, numLeads)
	ch <- prometheus.MustNewConstMetric(collector.numAnonymousMetric, prometheus.CounterValue, numAnonymous)
	ch <- prometheus.MustNewConstMetric(collector.numEmailsMetric, prometheus.CounterValue, numEmails)
	ch <- prometheus.MustNewConstMetric(collector.numCampaignsMetric, prometheus.CounterValue, numCampaigns)
	ch <- prometheus.MustNewConstMetric(collector.numPageHitsMetric, prometheus.CounterValue, numPageHits)
	ch <- prometheus.MustNewConstMetric(collector.numWebhookQueueMetric, prometheus.CounterValue, numWebhookQueue)
	ch <- prometheus.MustNewConstMetric(collector.numMessageQueueMetric, prometheus.CounterValue, numMessageQueue)

}

func main() {

	mtHostPtr := flag.String("host", "127.0.0.1", "Hostname or Address for DB server")
	mtPortPtr := flag.String("port", "3306", "DB server port")
	mtNamePtr := flag.String("db", "", "DB name")
	mtUserPtr := flag.String("user", "", "DB user for connection")
	mtPassPtr := flag.String("pass", "", "DB password for connection")
	mtTablePrefixPtr := flag.String("tableprefix", "", "Table prefix for Mautic tables")

	flag.Parse()

	dbHost := fmt.Sprintf("%s:%s", *mtHostPtr, *mtPortPtr)
	dbName := *mtNamePtr
	dbUser := *mtUserPtr
	dbPassword := *mtPassPtr
	tablePrefix := *mtTablePrefixPtr

	if dbName == "" {
		fmt.Fprintf(os.Stderr, "flag -db=dbname required!\n")
		os.Exit(1)
	}

	if dbUser == "" {
		fmt.Fprintf(os.Stderr, "flag -user=username required!\n")
		os.Exit(1)
	}

	//We create the collector
	collector := newMauticCollector(dbHost, dbName, dbUser, dbPassword, tablePrefix)
	prometheus.MustRegister(collector)

	//This section will start the HTTP server and expose
	//any metrics on the /metrics endpoint.
	http.Handle("/metrics", promhttp.Handler())
	log.Info("Beginning to serve on port :9117")
	log.Fatal(http.ListenAndServe(":9117", nil))
}
