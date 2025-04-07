# Remote NVIDIA vGPU Exporter

Prometheus-экспортер для сбора метрик NVIDIA vGPU. Собирает данные через SSH с хостов, 
оснащенных GPU NVIDIA, и предоставляет метрики в формате, совместимом с Prometheus.



## 📦 Особенности
- Сбор информации о vGPU: профиль, версия драйвера, VM
- Мониторинг памяти vGPU (Total/Used)
- Мониторинг утилизации GPU и памяти
- Минимальные расходы по ресурсам
- Docker-файл для удобного развертывания
- Интеграция с gpu-exporter [remote-nvidia-gpu-exporter](https://github.com/K1haa/remote-nvidia-gpu-exporter)
- Готовый дашборд Grafana (для совестной работы gpu и vgpu exporter)

## 📊 Метрики
| Метрика                             | Описание                          | Лейблы                                         |
|-------------------------------------|-----------------------------------|------------------------------------------------|
| `nvidia_vgpu_info`                  | Информация о vGPU                 | vgpu_id, vm_name, vgpu_profile, driver_version |
| `nvidia_vgpu_memory_bytes`          | Использование памяти vGPU         | vgpu_id, vm_name, type ("Total"/"Used")        |
| `nvidia_vgpu_utilization_percent`   | Утилизация vGPU (%)               | vgpu_id, vm_name, type ("Gpu"/"Memory")        |

## 🛠 Требования
- Go 1.22+ (только для сборки из исходников)
- Docker (для использования готового образа)
- Доступ по SSH к целевому хосту с NVIDIA GPU
- Prometheus + Grafana (для визуализации)

## 🐳 Быстрый старт
### Запуск через Docker [image](https://hub.docker.com/repository/docker/k1haa/nvidia-vgpu-exporter/general)
```bash
docker run -d \
  -e APP_USER="ваш_пользователь" \
  -e APP_HOST="ваш_хост" \
  -e APP_PASSWORD="ваш_пароль" \
  -p 9834:9834 \
  --name nvidia-vgpu-exporter \
  k1haa/nvidia-vgpu-exporter