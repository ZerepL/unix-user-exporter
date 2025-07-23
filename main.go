package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
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

func collectUserMetrics() {
	// Check if utmp file exists
	if _, err := os.Stat(*utmpPath); os.IsNotExist(err) {
		log.Printf("Error: utmp file %s does not exist", *utmpPath)
		return
	}

	// Try multiple commands to get user information
	var output []byte
	var err error
	var cmdOutput string

	// Try 'last' command with -f flag to specify utmp file
	cmd := exec.Command("last", "-f", *utmpPath, "-R")
	output, err = cmd.Output()
	if err == nil {
		cmdOutput = string(output)
		if *debugMode {
			log.Printf("Last command output: %s", cmdOutput)
		}
	} else if *debugMode {
		log.Printf("Error executing 'last -f %s -R' command: %v", *utmpPath, err)
	}

	// If 'last' fails or returns no useful data, try 'who'
	if err != nil || !strings.Contains(cmdOutput, "still logged in") {
		// Try 'who' command
		cmd = exec.Command("who")
		output, err = cmd.Output()
		if err == nil {
			cmdOutput = string(output)
			if *debugMode {
				log.Printf("Who command output: %s", cmdOutput)
			}
		} else if *debugMode {
			log.Printf("Error executing 'who' command: %v", err)
		}
	}

	// If both 'last' and 'who' fail, try 'w'
	if err != nil || cmdOutput == "" {
		// Try 'w' command
		cmd = exec.Command("w", "-h")
		output, err = cmd.Output()
		if err == nil {
			cmdOutput = string(output)
			if *debugMode {
				log.Printf("W command output: %s", cmdOutput)
			}
		} else if *debugMode {
			log.Printf("Error executing 'w -h' command: %v", err)
			return
		}
	}

	// Reset metrics before collecting new data
	userSessions.Reset()
	userSessionCount.Reset()
	userSessionByIP.Reset()

	// Parse the output
	lines := strings.Split(cmdOutput, "\n")
	
	// Count valid lines (users)
	validLines := 0
	userCounts := make(map[string]int)
	ipCounts := make(map[string]int)

	for _, line := range lines {
		if line == "" || strings.Contains(line, "wtmp begins") || !strings.Contains(line, "still logged in") {
			continue
		}
		
		// Split the line into fields
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		validLines++
		username := fields[0]
		tty := fields[1]
		
		// Handle the 'from' field which might contain IP or hostname
		from := "-"
		if len(fields) > 2 {
			from = fields[2]
		}
		
		// Get login time
		loginTime := "-"
		if len(fields) > 3 {
			loginTime = strings.Join(fields[3:5], " ")
		}

		// Increment user count
		userCounts[username]++

		// Increment IP count (if it's an IP address)
		if from != "-" && from != ":0" && from != ":0.0" {
			ipCounts[from]++
		}

		// Set user session info
		userSessions.WithLabelValues(username, from, tty, loginTime).Set(1)
	}

	// Set total users metric
	totalUsers.Set(float64(validLines))

	// Set user count metrics
	for username, count := range userCounts {
		userSessionCount.WithLabelValues(username).Set(float64(count))
	}

	// Set IP count metrics
	for ip, count := range ipCounts {
		userSessionByIP.WithLabelValues(ip).Set(float64(count))
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
