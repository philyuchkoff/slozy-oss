#!/usr/bin/env python3
"""
Пример создания SLO через Python API
"""

import requests
import json
import sys

BASE_URL = "http://localhost:8080"

def create_slo_with_alerting():
    """Создание SLO с алертингом"""
    
    slo_data = {
        "serviceName": "notification-service",
        "sloName": "api-latency",
        "sloDescription": "99th percentile latency for notification API",
        "indicatorType": "latency",
        "metricSource": "prometheus",
        "metricQuery": "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job=\"notification-service\"}[5m]))",
        "target": 99.9,
        "threshold": "500ms",
        "budgetingPeriod": "28d",
        "team": "backend",
        "labels": {
            "environment": "production",
            "tier": "critical"
        },
        "alerting": {
            "enabled": True,
            "notificationChannels": [
                "slack#notifications-alerts",
                "email:oncall@example.com"
            ],
            "burnRateThresholds": [
                {
                    "rate": 10.0,
                    "period": "1h",
                    "severity": "warning",
                    "alertName": "HighBurnRate"
                },
                {
                    "rate": 50.0,
                    "period": "5m",
                    "severity": "critical",
                    "alertName": "CriticalBurnRate"
                }
            ]
        }
    }
    
    try:
        print("Отправка запроса на создание SLO...")
        response = requests.post(
            f"{BASE_URL}/api/slo/create",
            json=slo_data,
            headers={"Content-Type": "application/json"},
            stream=True
        )
        
        if response.status_code == 200 and response.headers.get('content-type') == 'application/zip':
            filename = "notification-service-config.zip"
            
            # Получаем имя файла из заголовка
            content_disposition = response.headers.get('content-disposition', '')
            if 'filename=' in content_disposition:
                filename = content_disposition.split('filename=')[1].strip('"')
            
            # Сохраняем ZIP файл
            with open(filename, 'wb') as f:
                for chunk in response.iter_content(chunk_size=8192):
                    f.write(chunk)
            
            print(f"✅ SLO успешно создан и сохранен в {filename}")
            return True
            
        else:
            print(f"❌ Ошибка при создании SLO: {response.status_code}")
            print(f"Ответ: {response.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"❌ Ошибка соединения: {e}")
        return False

def list_existing_slos():
    """Получение списка существующих SLO"""
    
    try:
        response = requests.get(f"{BASE_URL}/api/slos")
        if response.status_code == 200:
            slos = response.json().get('slos', [])
            print(f"\n📋 Найдено {len(slos)} SLO:")
            for slo in slos:
                print(f"  • {slo['name']}")
            return slos
        else:
            print("❌ Ошибка при получении списка SLO")
            return []
    except Exception as e:
        print(f"❌ Ошибка: {e}")
        return []

def get_templates():
    """Получение доступных шаблонов"""
    
    try:
        response = requests.get(f"{BASE_URL}/api/templates")
        if response.status_code == 200:
            templates = response.json().get('templates', [])
            print(f"\n📚 Доступно {len(templates)} шаблонов:")
            for template in templates:
                print(f"  • {template['displayName']} - {template['description']}")
            return templates
        else:
            print("❌ Ошибка при получении шаблонов")
            return []
    except Exception as e:
        print(f"❌ Ошибка: {e}")
        return []

def main():
    """Основная функция"""
    
    print("=" * 60)
    print("SLOzy - Примеры использования API")
    print("=" * 60)
    
    # Проверяем здоровье сервиса
    try:
        health = requests.get(f"{BASE_URL}/health").json()
        print(f"✅ Сервис работает: {health.get('status')}")
    except:
        print("❌ Не удалось подключиться к сервису")
        sys.exit(1)
    
    # Получаем шаблоны
    templates = get_templates()
    
    # Получаем список существующих SLO
    slos = list_existing_slos()
    
    # Создаем новый SLO
    print("\n🔄 Создание нового SLO с алертингом...")
    if create_slo_with_alerting():
        print("\n🎉 Пример успешно выполнен!")
        print("\nСледующие шаги:")
        print("1. Распакуйте ZIP архив")
        print("2. Настройте метрики под ваше окружение")
        print("3. Импортируйте конфигурации в вашу систему мониторинга")
    else:
        print("\n❌ Не удалось создать SLO")

if __name__ == "__main__":
    main()