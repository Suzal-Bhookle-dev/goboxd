#!/bin/bash

set -e

echo "Installing Rust..."
apt-get install -y rustc

rustc --version
