#!/bin/bash

# Default port
PORT=32142

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --port=*)
      PORT="${1#*=}"
      shift
      ;;
    *)
      echo "Unknown option: $1"
      echo "Usage: $0 [--port=PORT]"
      exit 1
      ;;
  esac
done

# Function to generate metrics
generate_metrics() {
  # Get the output of the w command
  W_OUTPUT=$(w -h)
  
  # Count the number of lines (users)
  USER_COUNT=$(echo "$W_OUTPUT" | wc -l)
  
  # Initialize metrics
  METRICS="# HELP unix_users_logged_in_total Total number of users currently logged in\n"
  METRICS+="# TYPE unix_users_logged_in_total gauge\n"
  METRICS+="unix_users_logged_in_total $USER_COUNT\n\n"
  
  METRICS+="# HELP unix_user_session_info Information about user sessions\n"
  METRICS+="# TYPE unix_user_session_info gauge\n"
  
  METRICS+="# HELP unix_user_session_count Number of sessions per user\n"
  METRICS+="# TYPE unix_user_session_count gauge\n"
  
  METRICS+="# HELP unix_user_session_by_ip Number of sessions per origin IP\n"
  METRICS+="# TYPE unix_user_session_by_ip gauge\n"
  
  # Process each line of the w output
  declare -A user_counts
  declare -A ip_counts
  
  while IFS= read -r line; do
    if [ -z "$line" ]; then
      continue
    fi
    
    # Parse the line
    read -r username tty from login_time idle jcpu pcpu what <<< "$line"
    
    # Escape quotes in labels
    username=$(echo "$username" | sed 's/"/\\"/g')
    from=$(echo "$from" | sed 's/"/\\"/g')
    tty=$(echo "$tty" | sed 's/"/\\"/g')
    login_time=$(echo "$login_time" | sed 's/"/\\"/g')
    what=$(echo "$what" | sed 's/"/\\"/g')
    
    # Add user session info metric
    METRICS+="unix_user_session_info{username=\"$username\",from=\"$from\",tty=\"$tty\",login_time=\"$login_time\"} 1\n"
    
    # Count sessions per user
    user_counts["$username"]=$((${user_counts["$username"]:-0} + 1))
    
    # Count sessions per IP
    if [ "$from" != "-" ] && [ "$from" != ":0" ] && [ "$from" != ":0.0" ]; then
      ip_counts["$from"]=$((${ip_counts["$from"]:-0} + 1))
    fi
  done <<< "$W_OUTPUT"
  
  # Add user count metrics
  for user in "${!user_counts[@]}"; do
    METRICS+="unix_user_session_count{username=\"$user\"} ${user_counts[$user]}\n"
  done
  
  # Add IP count metrics
  for ip in "${!ip_counts[@]}"; do
    METRICS+="unix_user_session_by_ip{ip=\"$ip\"} ${ip_counts[$ip]}\n"
  done
  
  echo -e "$METRICS"
}

# Start a simple HTTP server
echo "Starting Unix User Exporter on port $PORT"
while true; do
  echo -e "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\n$(generate_metrics)" | nc -l -p $PORT
done
