package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"flag"
	"fmt"
	"os"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

const (
	namespace = "mautic"
)

//This is my collector metrics
type mauticCollector struct {
	numEmailsSentMetric      *prometheus.Desc
	numLeadsMetric           *prometheus.Desc
	numAnonymousMetric       *prometheus.Desc
	numEmailsMetric          *prometheus.Desc
	numCampaignsMetric       *prometheus.Desc
	numSegmentsMetric        *prometheus.Desc
	numPageHitsMetric        *prometheus.Desc
	numWebhookQueueMetric    *prometheus.Desc
	numMessageQueueMetric    *prometheus.GaugeVec
	numLeadEventFailedMetric *prometheus.Desc
	numCampaignEventsMetric  *prometheus.GaugeVec
	numNotificationsMetric   *prometheus.Desc
	numPageRedirectMetric    *prometheus.Desc
	numLeadsinCampaignMetric *prometheus.GaugeVec
	numLeadsinSegmentMetric  *prometheus.GaugeVec

	dbHost        string
	dbName        string
	dbUser        string
	dbPass        string
	dbTablePrefix string
}

//This is a constructor for my mauticCollector struct
func newMauticCollector(host string, dbname string, username string, pass string, tablePrefix string) *mauticCollector {
	return &mauticCollector{
		numEmailsSentMetric: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "emails_sent_total"),
			"Shows the number of emails sent in Mautic (AUTO_INCREMENT value)",
			nil, nil,
		),
		numLeadsMetric: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "leads_total"),
			"Shows the number of leads in Mautic",
			nil, nil,
		),
		numAnonymousMetric: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "anonymous_total"),
			"Shows the number of anonymous in Mautic",
			nil, nil,
		),
		numEmailsMetric: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "emails_total"),
			"Shows the number of emails in Mautic",
			nil, nil,
		),
		numCampaignsMetric: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "campaigns_total"),
			"Shows the number of campaigns in Mautic",
			nil, nil,
		),
		numSegmentsMetric: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "segments_total"),
			"Shows the number of segments in Mautic",
			nil, nil,
		),
		numPageHitsMetric: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "page_hits_total"),
			"Shows the number of page hits in Mautic",
			nil, nil,
		),
		numWebhookQueueMetric: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "webhook_total"),
			"Shows the number of webhook in Mautic's queue",
			nil, nil,
		),
		numMessageQueueMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "queue_total",
			Help:      "Shows the number of message in Mautic's queue",
		},
			[]string{"type"},
		),
		numCampaignEventsMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "campaign_events_total",
			Help:      "Shows the number of campaign events in Mautic",
		},
			[]string{"type"},
		),
		numLeadEventFailedMetric: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "lead_event_failed_total"),
			"Shows the number of failed events on leads",
			nil, nil,
		),
		numNotificationsMetric: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "notifications_total"),
			"Shows the number of notifications in Mautic",
			nil, nil,
		),
		numPageRedirectMetric: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "page_redirect_total"),
			"Shows the number of page redirects in Mautic",
			nil, nil,
		),
		numLeadsinCampaignMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "leads_in_campaign_total",
			Help:      "Shows the number of leads in a campaign",
		},
			[]string{"type"},
		),
		numLeadsinSegmentMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "leads_in_segment_total",
			Help:      "Shows the number of leads in a segment",
		},
			[]string{"type"},
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
	ch <- collector.numEmailsSentMetric
	ch <- collector.numLeadsMetric
	ch <- collector.numAnonymousMetric
	ch <- collector.numEmailsMetric
	ch <- collector.numCampaignsMetric
	ch <- collector.numSegmentsMetric
	ch <- collector.numPageHitsMetric
	ch <- collector.numWebhookQueueMetric
	ch <- collector.numLeadEventFailedMetric
	ch <- collector.numNotificationsMetric
	ch <- collector.numPageRedirectMetric
	collector.numMessageQueueMetric.Describe(ch)
	collector.numCampaignEventsMetric.Describe(ch)
	collector.numLeadsinCampaignMetric.Describe(ch)
	collector.numLeadsinSegmentMetric.Describe(ch)
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

	//SELECT AUTO_INCREMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA = "mautic" AND TABLE_NAME = "email_stats";
	queryNumEmailsSentMetric := fmt.Sprintf("SELECT AUTO_INCREMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA = %s AND TABLE_NAME = %semail_stats;", collector.dbName, collector.dbTablePrefix)
	mtQueryCounter(db, ch, collector.numEmailsSentMetric, queryNumEmailsSentMetric)

	//select count(id) as numLeads from leads where date_identified is not null
	queryNumLeadsMetric := fmt.Sprintf("select count(id) as numLeads from %sleads where date_identified is not null;", collector.dbTablePrefix)
	mtQueryGauge(db, ch, collector.numLeadsMetric, queryNumLeadsMetric)

	//select count(id) as numAnonymous from leads where date_identified is null
	queryNumAnonymousMetric := fmt.Sprintf("select count(id) as numAnonymous from %sleads where date_identified is null;", collector.dbTablePrefix)
	mtQueryGauge(db, ch, collector.numAnonymousMetric, queryNumAnonymousMetric)

	//select count(*) as numEmails from emails;
	queryNumEmailsMetric := fmt.Sprintf("select count(*) as numEmails from %semails;", collector.dbTablePrefix)
	mtQueryGauge(db, ch, collector.numEmailsMetric, queryNumEmailsMetric)

	//select count(*) as numCampaigns from campaigns;
	queryNumCampaignsMetric := fmt.Sprintf("select count(*) as numCampaigns from %scampaigns;", collector.dbTablePrefix)
	mtQueryGauge(db, ch, collector.numCampaignsMetric, queryNumCampaignsMetric)

	//select count(*) as numCampaigns from campaigns;
	queryNumSegmentsMetric := fmt.Sprintf("select count(*) as numSegments from %slead_lists;", collector.dbTablePrefix)
	mtQueryGauge(db, ch, collector.numSegmentsMetric, queryNumSegmentsMetric)

	//select count(*) as numPageHits from page_hits;
	queryNumPageHitsMetric := fmt.Sprintf("select count(*) as numPageHits from %spage_hits;", collector.dbTablePrefix)
	mtQueryCounter(db, ch, collector.numPageHitsMetric, queryNumPageHitsMetric)

	//select count(*) as numWebhookQueue from webhook_queue;
	queryNumWebhookQueueMetric := fmt.Sprintf("select count(*) as numWebhookQueue from %swebhook_queue;", collector.dbTablePrefix)
	mtQueryGauge(db, ch, collector.numWebhookQueueMetric, queryNumWebhookQueueMetric)

	//select status as label, count(*) as value from message_queue group by status;
	queryNumMessageQueueMetric := fmt.Sprintf("select status as label, count(*) as value from %smessage_queue group by status;", collector.dbTablePrefix)
	mtQueryGaugeVec(db, ch, collector.numMessageQueueMetric, queryNumMessageQueueMetric)

	//select type as label, count(*) as value from campaign_events GROUP BY type;
	queryNumCampaignEventsMetric := fmt.Sprintf("select type as label, count(*) as value from %scampaign_events group by type;", collector.dbTablePrefix)
	mtQueryGaugeVec(db, ch, collector.numCampaignEventsMetric, queryNumCampaignEventsMetric)

	//select count(*) from campaign_lead_event_failed_log;
	queryNumLeadEventFailedMetric := fmt.Sprintf("select count(*) from %scampaign_lead_event_failed_log;", collector.dbTablePrefix)
	mtQueryCounter(db, ch, collector.numLeadEventFailedMetric, queryNumLeadEventFailedMetric)

	//select count(*) from notifications;
	queryNumNotificationsMetric := fmt.Sprintf("select count(*) from %snotifications;", collector.dbTablePrefix)
	mtQueryGauge(db, ch, collector.numNotificationsMetric, queryNumNotificationsMetric)

	//select count(*) from page_redirects;
	queryNumPageRedirectMetric := fmt.Sprintf("select count(*) from %spage_redirects;", collector.dbTablePrefix)
	mtQueryCounter(db, ch, collector.numPageRedirectMetric, queryNumPageRedirectMetric)

	//select campaign_id as label, count(*) as value from campaign_leads where manually_removed = 0  group by campaign_id;
	queryNumLeadsinCampaignMetric := fmt.Sprintf("select campaign_id as label, count(*) as value from %scampaign_leads where manually_removed = 0 group by campaign_id;", collector.dbTablePrefix)
	mtQueryGaugeVec(db, ch, collector.numLeadsinCampaignMetric, queryNumLeadsinCampaignMetric)

	//select leadlist_id as label, count(*) as value from lead_lists_leads group by leadlist_id ;
	queryNumLeadsinSegmentMetric := fmt.Sprintf("select leadlist_id as label, count(*) as value from %slead_lists_leads group by leadlist_id;", collector.dbTablePrefix)
	mtQueryGaugeVec(db, ch, collector.numLeadsinSegmentMetric, queryNumLeadsinSegmentMetric)

}

func mtQueryCounter(db *sql.DB, ch chan<- prometheus.Metric, desc *prometheus.Desc, mysqlRequest string) {
	var value float64
	var err = db.QueryRow(mysqlRequest).Scan(&value)
	if err != nil {
		log.Fatal(err)
	}
	ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value)
}

func mtQueryGauge(db *sql.DB, ch chan<- prometheus.Metric, desc *prometheus.Desc, mysqlRequest string) {
	var value float64
	var err = db.QueryRow(mysqlRequest).Scan(&value)
	if err != nil {
		log.Fatal(err)
	}
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value)
}

func mtQueryGaugeVec(db *sql.DB, ch chan<- prometheus.Metric, desc *prometheus.GaugeVec, mysqlRequest string) {
	rows, err := db.Query(mysqlRequest)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	desc.Reset()
	for rows.Next() {
		var label string
		var value float64
		err = rows.Scan(&label, &value)
		if err != nil {
			log.Fatal(err)
		}
		desc.WithLabelValues(label).Set(value)
	}
	desc.Collect(ch)
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
	log.Info("Beginning to serve on port :9851")
	log.Fatal(http.ListenAndServe(":9851", nil))
}
