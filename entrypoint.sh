#!/bin/bash

# Экспорт переменных
export APP_USER=${USER:-root}
export APP_HOST=${HOST}         
export APP_PASSWORD=${PASS}      

exec nvidia-vgpu-exporter