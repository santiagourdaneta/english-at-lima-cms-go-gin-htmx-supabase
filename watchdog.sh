#!/bin/bash
echo "👀 Vigilando el servidor..."
while true; do
  # Muestra el proceso de Go y cuánta memoria usa
  ps aux | grep "main" | grep -v "grep" | awk '{print "RAM Uso: " $4 "% | CPU Uso: " $3 "%"}'
  sleep 2
done