#!/bin/bash

# Function to get hardware info on Linux
get_linux_info() {
    echo "Operating System: Linux"
    echo "CPU Architecture: $(uname -m)"
    echo "Number of CPUs: $(grep -c ^processor /proc/cpuinfo)"
    echo "Total RAM: $(grep MemTotal /proc/meminfo | awk '{print $2 / 1024 / 1024 " GB"}')"
}

# Function to get hardware info on macOS
get_mac_info() {
    echo "Operating System: macOS"
    echo "CPU Architecture: $(uname -m)"
    echo "Number of CPUs: $(sysctl -n hw.ncpu)"
    total_ram_bytes=$(sysctl -n hw.memsize)
    total_ram_gb=$(echo "scale=2; $total_ram_bytes / 1024 / 1024 / 1024" | bc)
    echo "Total RAM: ${total_ram_gb} GB"
}

# Check the operating system and call the appropriate function
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    get_linux_info
elif [[ "$OSTYPE" == "darwin"* ]]; then
    get_mac_info
else
    echo "Unsupported operating system: $OSTYPE"
fi

