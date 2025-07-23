package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listenAddress = flag.String("web.listen-address", ":32142", "Address to listen on for web interface and telemetry")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics")
	debugMode     = flag.Bool("debug", false, "Enable debug logging")
	utmpPath      = flag.String("utmp-path", "/var/run/utmp", "Path to the utmp file")
)

// utmp entry types
const (
	EMPTY         = 0
	RUN_LVL       = 1
	BOOT_TIME     = 2
	NEW_TIME      = 3
	OLD_TIME      = 4
	INIT_PROCESS  = 5
	LOGIN_PROCESS = 6
	USER_PROCESS  = 7
	DEAD_PROCESS  = 8
	ACCOUNTING    = 9
)

// Define metrics
var (
	// Gauge for total number of logged-in users
	totalUsers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "unix_users_logged_in_total",
		Help: "Total number of users currently logged in",
	})

	// Gauge vector for user sessions with labels
	userSessions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "unix_user_session_info",
			Help: "Information about user sessions",
		},
		[]string{"username", "from", "tty", "login_time"},
	)

	// Gauge for user session count by username
	userSessionCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "unix_user_session_count",
			Help: "Number of sessions per user",
		},
		[]string{"username"},
	)

	// Gauge for user session count by origin IP
	userSessionByIP = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "unix_user_session_by_ip",
			Help: "Number of sessions per origin IP",
		},
		[]string{"ip"},
	)
)

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(totalUsers)
	prometheus.MustRegister(userSessions)
	prometheus.MustRegister(userSessionCount)
	prometheus.MustRegister(userSessionByIP)
}

func parseUtmpEntry(data []byte) (int32, int32, string, string, string, time.Time) {
	if len(data) < 384 {
		return 0, 0, "", "", "", time.Time{}
	}

	// Parse fields manually from byte array
	entryType := int32(data[0]) | int32(data[1])<<8 | int32(data[2])<<16 | int32(data[3])<<24
	pid := int32(data[4]) | int32(data[5])<<8 | int32(data[6])<<16 | int32(data[7])<<24
	
	// Line starts at offset 8, 32 bytes
	line := strings.TrimRight(string(data[8:40]), "\x00")
	
	// User starts at offset 44, 32 bytes  
	user := strings.TrimRight(string(data[44:76]), "\x00")
	
	// Host starts at offset 76, 256 bytes
	host := strings.TrimRight(string(data[76:332]), "\x00")
	
	// Time starts at offset 340, 8 bytes (2 int32s)
	timeSec := int32(data[340]) | int32(data[341])<<8 | int32(data[342])<<16 | int32(data[343])<<24
	loginTime := time.Unix(int64(timeSec), 0)
	
	return entryType, pid, user, line, host, loginTime
}

func collectUserMetrics() {
	// Check if utmp file exists
	if _, err := os.Stat(*utmpPath); os.IsNotExist(err) {
		log.Printf("Error: utmp file %s does not exist", *utmpPath)
		return
	}

	// Open and read utmp file
	file, err := os.Open(*utmpPath)
	if err != nil {
		log.Printf("Error opening utmp file %s: %v", *utmpPath, err)
		return
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("Error getting file info: %v", err)
		return
	}

	entrySize := 384 // Standard Linux utmp entry size
	fileSize := fileInfo.Size()
	numEntries := int(fileSize) / entrySize

	if *debugMode {
		log.Printf("utmp file size: %d bytes, entry size: %d bytes, entries: %d", 
			fileSize, entrySize, numEntries)
	}

	// Reset metrics before collecting new data
	userSessions.Reset()
	userSessionCount.Reset()
	userSessionByIP.Reset()

	// Count valid sessions
	validSessions := 0
	userCounts := make(map[string]int)
	ipCounts := make(map[string]int)

	// Read all data at once
	data := make([]byte, fileSize)
	_, err = file.Read(data)
	if err != nil {
		log.Printf("Error reading utmp file: %v", err)
		return
	}

	// Parse each entry
	for i := 0; i < numEntries; i++ {
		offset := i * entrySize
		if offset+entrySize > len(data) {
			break
		}

		entryData := data[offset : offset+entrySize]
		entryType, pid, username, tty, host, loginTime := parseUtmpEntry(entryData)

		if *debugMode {
			log.Printf("Entry %d: type=%d, pid=%d, user='%s', line='%s', host='%s'", 
				i, entryType, pid, username, tty, host)
		}

		// Only process USER_PROCESS entries (active user sessions)
		if entryType != USER_PROCESS {
			continue
		}

		// Skip empty usernames
		if username == "" {
			continue
		}

		loginTimeStr := loginTime.Format("2006-01-02 15:04")

		if *debugMode {
			log.Printf("Found user session: user=%s, tty=%s, host=%s, time=%s", 
				username, tty, host, loginTimeStr)
		}

		validSessions++

		// Increment user count
		userCounts[username]++

		// Increment IP count (if it's not empty and not local)
		if host != "" && host != ":0" && host != ":0.0" && host != "console" {
			ipCounts[host]++
		}

		// Set user session info
		from := host
		if from == "" {
			from = "-"
		}
		userSessions.WithLabelValues(username, from, tty, loginTimeStr).Set(1)
	}

	// Set total users metric
	totalUsers.Set(float64(validSessions))

	// Set user count metrics
	for username, count := range userCounts {
		userSessionCount.WithLabelValues(username).Set(float64(count))
	}

	// Set IP count metrics
	for ip, count := range ipCounts {
		userSessionByIP.WithLabelValues(ip).Set(float64(count))
	}

	if *debugMode {
		log.Printf("Collected metrics: %d total sessions, %d unique users, %d unique IPs", 
			validSessions, len(userCounts), len(ipCounts))
	}
}

func main() {
	flag.Parse()

	if *debugMode {
		log.Println("Debug mode enabled")
		log.Printf("Using utmp file: %s", *utmpPath)
	}

	// Initial collection
	collectUserMetrics()

	// Start a goroutine to periodically collect metrics
	go func() {
		for {
			time.Sleep(15 * time.Second)
			collectUserMetrics()
		}
	}()

	// Set up HTTP server
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Unix User Exporter</title></head>
			<body>
			<h1>Unix User Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Printf("Server listening on %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
