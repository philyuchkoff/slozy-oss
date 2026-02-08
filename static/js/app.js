// Tab management
function showTab(tabName) {
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.remove('active');
    });
    document.querySelectorAll('.tab').forEach(tab => {
        tab.classList.remove('active');
    });

    document.getElementById(tabName + '-tab').classList.add('active');
    var tabElements = document.querySelectorAll('.tab');
    for (var i = 0; i < tabElements.length; i++) {
        if (tabElements[i].getAttribute('onclick') === "showTab('" + tabName + "')") {
            tabElements[i].classList.add('active');
            break;
        }
    }

    if (tabName === 'templates') {
        loadTemplates();
    } else if (tabName === 'existing') {
        loadExistingSLOs();
    }
}

// Sync range and number inputs for target
document.getElementById('targetRange').addEventListener('input', function () {
    document.getElementById('target').value = this.value;
});

document.getElementById('target').addEventListener('input', function () {
    document.getElementById('targetRange').value = this.value;
});

// Label management
function addLabel() {
    const container = document.getElementById('labelsContainer');
    const div = document.createElement('div');
    div.className = 'label-input';
    div.innerHTML = `
        <input type="text" placeholder="key" class="label-key">
        <input type="text" placeholder="value" class="label-value">
        <button type="button" onclick="removeLabel(this)" class="btn btn-secondary">
            <i class="fas fa-times"></i>
        </button>
    `;
    container.appendChild(div);
}

function removeLabel(button) {
    button.parentElement.remove();
}

// Alerting functions
function toggleAlertingConfig() {
    const enabled = document.getElementById('enableAlerting').checked;
    document.getElementById('alertingConfig').style.display = enabled ? 'block' : 'none';
}

function addThreshold() {
    const container = document.getElementById('burnRateThresholds');
    const div = document.createElement('div');
    div.className = 'threshold-row';

    div.innerHTML = `
        <select class="threshold-severity">
            <option value="warning">Warning</option>
            <option value="critical">Critical</option>
        </select>
        <input type="number" class="threshold-rate" placeholder="10" step="0.1" min="1" value="10">
        <span class="threshold-label">x faster for</span>
        <select class="threshold-period">
            <option value="5m">5 minutes</option>
            <option value="30m">30 minutes</option>
            <option value="1h" selected>1 hour</option>
            <option value="6h">6 hours</option>
            <option value="1d">1 day</option>
        </select>
        <input type="text" class="threshold-name" placeholder="Alert name" value="HighBurnRate">
        <button type="button" onclick="removeThreshold(this)" class="btn btn-secondary btn-danger">
            <i class="fas fa-times"></i>
        </button>
    `;

    container.appendChild(div);
}

function removeThreshold(button) {
    if (document.querySelectorAll('.threshold-row').length > 1) {
        button.parentElement.remove();
    } else {
        alert('At least one threshold is required when alerting is enabled.');
    }
}

// Example queries
function setExample(type) {
    const examples = {
        latency: 'histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job="my-service"}[5m]))',
        availability: 'sum(rate(http_requests_total{job="my-service", status=~"2.."}[5m])) / sum(rate(http_requests_total{job="my-service"}[5m]))',
        throughput: 'rate(http_requests_total{job="my-service"}[5m])'
    };
    document.getElementById('metricQuery').value = examples[type] || '';
}

// Fill form with example data
function fillExample() {
    document.getElementById('serviceName').value = 'user-api';
    document.getElementById('sloName').value = 'api-latency';
    document.getElementById('sloDescription').value = '99th percentile latency for API endpoints';
    document.getElementById('team').value = 'backend';
    document.getElementById('indicatorType').value = 'latency';
    document.getElementById('target').value = 99.9;
    document.getElementById('targetRange').value = 99.9;
    document.getElementById('threshold').value = '200ms';
    document.getElementById('budgetingPeriod').value = '28d';
    document.getElementById('metricSource').value = 'prometheus';
    document.getElementById('metricQuery').value = 'histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job="user-api"}[5m]))';
    document.getElementById('serviceDescription').value = 'User management API service';

    // Enable and configure alerting
    document.getElementById('enableAlerting').checked = true;
    toggleAlertingConfig();
    document.getElementById('notificationChannels').value = 'slack#slo-alerts, email:sre@example.com, telegram:-1001234567890';

    showNotification('Example data loaded with alerting!', 'success');
}

// Clear form
function clearForm() {
    if (confirm('Clear all form data?')) {
        document.getElementById('serviceName').value = '';
        document.getElementById('sloName').value = '';
        document.getElementById('sloDescription').value = '';
        document.getElementById('team').value = '';
        document.getElementById('indicatorType').value = 'latency';
        document.getElementById('target').value = 99.9;
        document.getElementById('targetRange').value = 99.9;
        document.getElementById('threshold').value = '200ms';
        document.getElementById('budgetingPeriod').value = '28d';
        document.getElementById('metricSource').value = 'prometheus';
        document.getElementById('metricQuery').value = '';
        document.getElementById('serviceDescription').value = '';
        document.getElementById('labelsContainer').innerHTML = `
            <div class="label-input">
                <input type="text" placeholder="key" class="label-key">
                <input type="text" placeholder="value" class="label-value">
                <button type="button" onclick="removeLabel(this)" class="btn btn-secondary">
                    <i class="fas fa-times"></i>
                </button>
            </div>
        `;

        // Clear alerting
        document.getElementById('enableAlerting').checked = false;
        toggleAlertingConfig();
        document.getElementById('notificationChannels').value = '';
        document.getElementById('burnRateThresholds').innerHTML = `
            <div class="threshold-row">
                <select class="threshold-severity">
                    <option value="warning">Warning</option>
                    <option value="critical">Critical</option>
                </select>
                <input type="number" class="threshold-rate" placeholder="10" step="0.1" min="1" value="10">
                <span class="threshold-label">x faster for</span>
                <select class="threshold-period">
                    <option value="5m">5 minutes</option>
                    <option value="30m">30 minutes</option>
                    <option value="1h" selected>1 hour</option>
                    <option value="6h">6 hours</option>
                    <option value="1d">1 day</option>
                </select>
                <input type="text" class="threshold-name" placeholder="Alert name" value="HighBurnRate">
                <button type="button" onclick="removeThreshold(this)" class="btn btn-secondary btn-danger">
                    <i class="fas fa-times"></i>
                </button>
            </div>
        `;

        showNotification('Form cleared!', 'success');
    }
}

// Load templates
async function loadTemplates() {
    try {
        const response = await fetch('/api/templates');
        const result = await response.json();

        const container = document.getElementById('templatesList');
        if (!container) {
            console.error('Templates container not found');
            return;
        }

        container.innerHTML = '';

        if (!result.templates || result.templates.length === 0) {
            container.innerHTML = `
                <div style="text-align: center; padding: 40px; color: #6c757d;">
                    <i class="fas fa-inbox fa-3x" style="margin-bottom: 20px;"></i>
                    <h4>No templates found</h4>
                    <p>Create your first template or add template files to the templates/ directory.</p>
                    <button onclick="createDefaultTemplate()" class="btn btn-primary">
                        <i class="fas fa-plus"></i> Create Default Templates
                    </button>
                </div>
            `;
            return;
        }

        result.templates.forEach(template => {
            const card = document.createElement('div');
            card.className = 'template-card';
            card.onclick = () => applyTemplate(template);

            let badges = '';
            if (template.defaults && template.defaults.alerting && template.defaults.alerting.enabled) {
                badges += '<span class="badge badge-success"><i class="fas fa-bell"></i> Alerts</span> ';
            }

            // Определяем иконку по типу сервиса
            let icon = 'fas fa-server';
            let serviceType = template.serviceType || (template.defaults && template.defaults.serviceType) || 'general';

            const iconMap = {
                'database': 'fas fa-database',
                'infrastructure': 'fas fa-cloud',
                'messaging': 'fas fa-exchange-alt',
                'web': 'fas fa-globe',
                'cache': 'fas fa-bolt',
                'storage': 'fas fa-hdd',
                'networking': 'fas fa-network-wired',
                'monitoring': 'fas fa-chart-line',
                'pipeline': 'fas fa-code-branch',
                'mobile': 'fas fa-mobile-alt',
                'iot': 'fas fa-microchip',
                'blockchain': 'fas fa-link'
            };

            icon = iconMap[serviceType] || 'fas fa-cogs';

            // Определяем цвет бейджа по типу сервиса
            const badgeClassMap = {
                'database': 'badge-primary',
                'infrastructure': 'badge-info',
                'messaging': 'badge-warning',
                'web': 'badge-success',
                'cache': 'badge-secondary',
                'storage': 'badge-dark',
                'networking': 'badge-primary',
                'monitoring': 'badge-info',
                'pipeline': 'badge-success',
                'mobile': 'badge-warning',
                'iot': 'badge-secondary',
                'blockchain': 'badge-dark'
            };

            const badgeClass = badgeClassMap[serviceType] || 'badge-secondary';
            badges += `<span class="badge ${badgeClass}">${serviceType}</span>`;

            const target = template.defaults && template.defaults.target ? template.defaults.target : 99.9;
            const indicatorType = template.defaults && template.defaults.indicatorType ? template.defaults.indicatorType : 'latency';

            card.innerHTML = `
                <div style="display: flex; align-items: start; gap: 15px; margin-bottom: 15px;">
                    <div style="flex-shrink: 0;">
                        <i class="${icon} fa-2x" style="color: #667eea;"></i>
                    </div>
                    <div style="flex: 1;">
                        <h4 style="margin: 0 0 5px 0; color: #495057;">${template.displayName || template.name}</h4>
                        <p style="margin: 0; color: #6c757d; font-size: 0.9em;">${template.description || ''}</p>
                        <div style="margin-top: 10px;">${badges}</div>
                        <div style="margin-top: 8px; font-size: 0.85em; color: #28a745;">
                            <i class="fas fa-bullseye"></i> Target: ${target}%
                        </div>
                        <div style="margin-top: 5px; font-size: 0.85em; color: #6c757d;">
                            <i class="fas fa-ruler"></i> ${indicatorType.charAt(0).toUpperCase() + indicatorType.slice(1)}
                        </div>
                    </div>
                    <div style="flex-shrink: 0; align-self: center;">
                        <button onclick="event.stopPropagation(); applyTemplate(${JSON.stringify(template).replace(/"/g, '&quot;')})" 
                                class="btn btn-primary btn-sm">
                            <i class="fas fa-check"></i> Use
                        </button>
                    </div>
                </div>
            `;

            container.appendChild(card);
        });

    } catch (error) {
        console.error('Failed to load templates:', error);
        showNotification('Failed to load templates', 'error');
    }
}

// Apply template
function applyTemplate(template) {
    const defaults = template.defaults || {};

    // Заполняем поля формы
    if (defaults.indicatorType) {
        document.getElementById('indicatorType').value = defaults.indicatorType;
    }
    if (defaults.target) {
        document.getElementById('target').value = defaults.target;
        document.getElementById('targetRange').value = defaults.target;
    }
    if (defaults.threshold) {
        document.getElementById('threshold').value = defaults.threshold;
    }
    if (defaults.budgetingPeriod) {
        document.getElementById('budgetingPeriod').value = defaults.budgetingPeriod;
    }
    if (defaults.metricSource) {
        document.getElementById('metricSource').value = defaults.metricSource;
    }
    if (defaults.metricQuery) {
        document.getElementById('metricQuery').value = defaults.metricQuery;
    }

    // Применяем алертинг если есть
    if (defaults.alerting) {
        document.getElementById('enableAlerting').checked = defaults.alerting.enabled || false;
        toggleAlertingConfig();

        if (defaults.alerting.notificationChannels) {
            document.getElementById('notificationChannels').value =
                Array.isArray(defaults.alerting.notificationChannels)
                    ? defaults.alerting.notificationChannels.join(', ')
                    : defaults.alerting.notificationChannels;
        }

        // Очищаем существующие thresholds
        const burnRateContainer = document.getElementById('burnRateThresholds');
        if (burnRateContainer) {
            burnRateContainer.innerHTML = '';

            // Добавляем thresholds из шаблона
            if (defaults.alerting.burnRateThresholds && Array.isArray(defaults.alerting.burnRateThresholds)) {
                defaults.alerting.burnRateThresholds.forEach(threshold => {
                    addThreshold();
                    const rows = document.querySelectorAll('.threshold-row');
                    const lastRow = rows[rows.length - 1];
                    if (lastRow) {
                        if (threshold.severity) {
                            lastRow.querySelector('.threshold-severity').value = threshold.severity;
                        }
                        if (threshold.rate) {
                            lastRow.querySelector('.threshold-rate').value = threshold.rate;
                        }
                        if (threshold.period) {
                            lastRow.querySelector('.threshold-period').value = threshold.period;
                        }
                        if (threshold.alertName) {
                            lastRow.querySelector('.threshold-name').value = threshold.alertName;
                        }
                    }
                });
            }

            // Если нет thresholds, добавляем один по умолчанию
            if (!defaults.alerting.burnRateThresholds || defaults.alerting.burnRateThresholds.length === 0) {
                addThreshold();
            }
        }
    } else {
        // Если алертинг не настроен, выключаем его
        document.getElementById('enableAlerting').checked = false;
        toggleAlertingConfig();
    }

    // Переключаемся на вкладку Create
    showTab('create');

    // Показываем уведомление
    showNotification(`Template "${template.displayName || template.name}" applied!`, 'success');

    // Скроллим к форме
    document.getElementById('create-tab').scrollIntoView({ behavior: 'smooth' });
}

// Load existing SLOs
async function loadExistingSLOs() {
    try {
        const response = await fetch('/api/slos');
        const result = await response.json();

        const container = document.getElementById('existingSLOsList');

        if (result.slos && result.slos.length > 0) {
            let html = '<div class="slos-grid">';

            result.slos.forEach(slo => {
                html += `
                    <div class="slo-card">
                        <div class="slo-card-header">
                            <strong>${slo.name}</strong>
                            <button onclick="viewSLO('${slo.name}')" class="btn btn-primary">
                                <i class="fas fa-eye"></i> View
                            </button>
                        </div>
                        <div class="slo-card-body">
                            <i class="fas fa-file-alt"></i> ${slo.name}
                        </div>
                        <a href="${slo.path}" target="_blank" class="slo-card-link">
                            <i class="fas fa-external-link-alt"></i> Open file
                        </a>
                    </div>
                `;
            });

            html += '</div>';
            container.innerHTML = html;
        } else {
            container.innerHTML = '<p class="no-slos">No SLOs created yet. Start by creating one!</p>';
        }
    } catch (error) {
        console.error('Failed to load SLOs:', error);
        container.innerHTML = '<p class="error-message">Failed to load SLOs</p>';
    }
}

// View SLO
async function viewSLO(filename) {
    try {
        const response = await fetch('/data/' + filename);
        const yaml = await response.text();

        document.getElementById('yamlOutput').textContent = yaml;
        document.getElementById('result').style.display = 'block';
        document.getElementById('message').innerHTML = `
            <div class="success">
                <i class="fas fa-check-circle"></i> Loaded: ${filename}
            </div>
        `;
        document.getElementById('filesList').innerHTML = `<p><strong>Main file:</strong> ${filename}</p>`;

        document.getElementById('result').scrollIntoView({ behavior: 'smooth' });
    } catch (error) {
        showNotification('Failed to load SLO: ' + error.message, 'error');
    }
}

// Create SLO
async function createSLO() {
    const submitBtn = document.getElementById('submitBtn');
    const originalHtml = submitBtn.innerHTML;
    submitBtn.innerHTML = '<div class="loading"></div> Generating SLO...';
    submitBtn.disabled = true;

    // Показываем индикатор загрузки
    document.getElementById('loadingIndicator').style.display = 'block';
    document.getElementById('result').style.display = 'block';
    document.getElementById('message').innerHTML = `
        <div class="success">
            <i class="fas fa-spinner fa-spin"></i> Creating SLO configuration...
        </div>
    `;

    // Collect labels
    const labels = {};
    document.querySelectorAll('.label-input').forEach(div => {
        const key = div.querySelector('.label-key').value.trim();
        const value = div.querySelector('.label-value').value.trim();
        if (key && value) {
            labels[key] = value;
        }
    });

    // Collect alerting thresholds
    const burnRateThresholds = [];
    document.querySelectorAll('.threshold-row').forEach(row => {
        const severity = row.querySelector('.threshold-severity').value;
        const rate = parseFloat(row.querySelector('.threshold-rate').value);
        const period = row.querySelector('.threshold-period').value;
        const alertName = row.querySelector('.threshold-name').value.trim() || 'HighBurnRate';

        if (rate && period) {
            burnRateThresholds.push({
                rate: rate,
                period: period,
                severity: severity,
                alertName: alertName,
            });
        }
    });

    // Валидация каналов нотификаций
    const channels = document.getElementById('notificationChannels').value
        .split(',')
        .map(c => c.trim())
        .filter(c => c.length > 0);

    for (const channel of channels) {
        if (!validateTelegramChannel(channel)) {
            submitBtn.innerHTML = originalHtml;
            submitBtn.disabled = false;
            document.getElementById('loadingIndicator').style.display = 'none';
            return;
        }
    }

    // Prepare data
    const sloData = {
        serviceName: document.getElementById('serviceName').value.trim(),
        serviceDescription: document.getElementById('serviceDescription').value.trim(),
        sloName: document.getElementById('sloName').value.trim(),
        sloDescription: document.getElementById('sloDescription').value.trim(),
        indicatorType: document.getElementById('indicatorType').value,
        metricSource: document.getElementById('metricSource').value,
        metricQuery: document.getElementById('metricQuery').value.trim(),
        target: parseFloat(document.getElementById('target').value),
        threshold: document.getElementById('threshold').value.trim(),
        budgetingPeriod: document.getElementById('budgetingPeriod').value,
        team: document.getElementById('team').value.trim() || 'unknown',
        labels: labels,
        alerting: {
            enabled: document.getElementById('enableAlerting').checked,
            notificationChannels: channels,
            burnRateThresholds: burnRateThresholds,
        }
    };

    try {
        const response = await fetch('/api/slo/create', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify(sloData)
        });

        // Проверяем, является ли ответ ZIP архивом
        const contentType = response.headers.get('content-type');

        if (contentType && contentType.includes('application/zip')) {
            // Получаем имя файла из заголовка
            const contentDisposition = response.headers.get('content-disposition');
            let filename = 'slo-configuration.zip';

            if (contentDisposition) {
                const matches = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/.exec(contentDisposition);
                if (matches != null && matches[1]) {
                    filename = matches[1].replace(/['"]/g, '');
                }
            }

            // Скачиваем ZIP архив
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = filename;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);

            // Обновляем информацию на странице
            document.getElementById('loadingIndicator').style.display = 'none';
            document.getElementById('message').innerHTML = `
                <div class="success">
                    <i class="fas fa-check-circle"></i> SLO created successfully!
                    <div class="success-details">
                        All configuration files have been downloaded as ZIP archive.
                    </div>
                </div>
            `;

            // Показываем содержимое основного файла
            const mainContent = `# Main SLO Configuration
service: ${sloData.serviceName}
slo: ${sloData.sloName}
target: ${sloData.target}%
indicator: ${sloData.indicatorType}

# Files included in the ZIP archive:
# - ${sloData.serviceName}-${sloData.sloName}.yaml (main SLO definition)`;

            if (sloData.alerting.enabled) {
                document.getElementById('yamlOutput').textContent = mainContent + `
# - ${sloData.serviceName}-${sloData.sloName}-prometheus-alerts.yaml (Prometheus alert rules)
# - ${sloData.serviceName}-${sloData.sloName}-grafana-dashboard.json (Grafana dashboard)
# - ${sloData.serviceName}-${sloData.sloName}-README.md (documentation)`;
            } else {
                document.getElementById('yamlOutput').textContent = mainContent;
            }

            // Показываем список файлов
            let filesList = `<p><strong>Files included in ZIP archive:</strong></p><ul>`;
            filesList += `<li>${sloData.serviceName}-${sloData.sloName}.yaml (main SLO)</li>`;

            if (sloData.alerting.enabled) {
                filesList += `<li>${sloData.serviceName}-${sloData.sloName}-prometheus-alerts.yaml</li>`;
                filesList += `<li>${sloData.serviceName}-${sloData.sloName}-grafana-dashboard.json</li>`;
                filesList += `<li>${sloData.serviceName}-${sloData.sloName}-README.md</li>`;
            }

            filesList += `</ul>`;
            document.getElementById('filesList').innerHTML = filesList;

            showNotification(`✅ All files downloaded as "${filename}"`, 'success');
            loadExistingSLOs();
        } else {
            // Обрабатываем JSON ответ (старый формат для совместимости)
            const result = await response.json();

            document.getElementById('loadingIndicator').style.display = 'none';

            if (result.success) {
                document.getElementById('yamlOutput').textContent = result.yaml;

                // Show list of generated files
                let filesHtml = '<p><strong>Generated files:</strong></p><ul>';
                filesHtml += '<li>' + result.filename + ' (main SLO)</li>';

                if (result.additionalFiles) {
                    for (const [key, filename] of Object.entries(result.additionalFiles)) {
                        if (key !== 'prometheusAlertsYaml' && key !== 'grafanaDashboardJson') {
                            filesHtml += '<li>' + filename + '</li>';
                        }
                    }
                }

                filesHtml += '</ul>';
                document.getElementById('filesList').innerHTML = filesHtml;

                document.getElementById('message').innerHTML = `
                    <div class="success">
                        <i class="fas fa-check-circle"></i> SLO created successfully!
                        <div class="success-details">
                            Service: ${result.service} | SLO: ${result.slo}
                        </div>
                    </div>
                `;

                document.getElementById('result').scrollIntoView({ behavior: 'smooth' });
                loadExistingSLOs();
                showNotification('SLO created successfully!', 'success');
            } else {
                let errorMessage = 'Unknown error';
                if (result.errors && result.errors.length > 0) {
                    errorMessage = result.errors.join(', ');
                } else if (result.error) {
                    errorMessage = result.error;
                }

                document.getElementById('message').innerHTML = `
                    <div class="error">
                        <i class="fas fa-exclamation-triangle"></i> ${errorMessage}
                    </div>
                `;
                showNotification('Failed to create SLO: ' + errorMessage, 'error');
            }
        }
    } catch (error) {
        console.error('Error:', error);
        document.getElementById('loadingIndicator').style.display = 'none';
        document.getElementById('message').innerHTML = `
            <div class="error">
                <i class="fas fa-exclamation-triangle"></i> Network error: ${error.message}
            </div>
        `;
        showNotification('Network error: ' + error.message, 'error');
    } finally {
        submitBtn.innerHTML = originalHtml;
        submitBtn.disabled = false;
    }
}

// Utility functions
function copyToClipboard() {
    const yaml = document.getElementById('yamlOutput').textContent;
    navigator.clipboard.writeText(yaml).then(() => {
        showNotification('YAML copied to clipboard!', 'success');
    }).catch(err => {
        showNotification('Failed to copy: ' + err.message, 'error');
    });
}

function clearResult() {
    document.getElementById('result').style.display = 'none';
    document.getElementById('loadingIndicator').style.display = 'none';
}

// Функция проверки формата Telegram chat ID
function validateTelegramChannel(channel) {
    if (channel.startsWith('telegram:')) {
        const chatId = channel.replace('telegram:', '').trim();
        // Проверка формата chat ID Telegram
        // Личные чаты: 123456789 (только цифры)
        // Группы/каналы: -1001234567890 (начинается с -100)
        if (!/^(-100\d+|\d+)$/.test(chatId)) {
            showNotification('⚠️ Неверный формат Telegram chat ID. Используйте: telegram:123456789 или telegram:-1001234567890', 'error');
            return false;
        }
    }
    return true;
}

// Функция для создания дефолтных шаблонов
async function createDefaultTemplate() {
    try {
        const response = await fetch('/api/templates/create-defaults', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            }
        });

        const result = await response.json();

        if (result.success) {
            showNotification('Default templates created successfully!', 'success');
            loadTemplates(); // Перезагружаем список
        } else {
            showNotification('Failed to create default templates: ' + (result.error || 'Unknown error'), 'error');
        }
    } catch (error) {
        console.error('Error creating default templates:', error);
        showNotification('Network error: ' + error.message, 'error');
    }
}

function showNotification(message, type) {
    const notification = document.createElement('div');
    notification.className = 'notification notification-' + type;

    notification.innerHTML = `
        <div class="notification-content">
            <span>${message}</span>
            <button onclick="this.parentElement.parentElement.remove()">
                <i class="fas fa-times"></i>
            </button>
        </div>
    `;

    document.body.appendChild(notification);

    setTimeout(() => {
        if (notification.parentElement) {
            notification.remove();
        }
    }, 5000);
}

// Add CSS for notifications
const notificationStyle = document.createElement('style');
notificationStyle.textContent = `
    .notification {
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 15px 20px;
        border-radius: 8px;
        color: white;
        font-weight: 600;
        z-index: 10000;
        animation: slideIn 0.3s ease;
        box-shadow: 0 5px 15px rgba(0,0,0,0.2);
        min-width: 300px;
    }
    
    .notification-success {
        background: linear-gradient(135deg, #28a745 0%, #20c997 100%);
    }
    
    .notification-error {
        background: linear-gradient(135deg, #dc3545 0%, #e83e8c 100%);
    }
    
    .notification-info {
        background: linear-gradient(135deg, #17a2b8 0%, #20c997 100%);
    }
    
    .notification-content {
        display: flex;
        align-items: center;
        justify-content: space-between;
    }
    
    .notification-content button {
        background: none;
        border: none;
        color: white;
        cursor: pointer;
        margin-left: 15px;
        font-size: 16px;
    }
    
    @keyframes slideIn {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }
    
    .slos-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
        gap: 15px;
    }
    
    .slo-card {
        background: #f8f9fa;
        padding: 15px;
        border-radius: 8px;
        border: 1px solid #e9ecef;
    }
    
    .slo-card-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 10px;
    }
    
    .slo-card-header .btn {
        padding: 5px 10px;
        font-size: 0.8em;
    }
    
    .slo-card-body {
        font-size: 0.85em;
        color: #6c757d;
        margin-bottom: 10px;
    }
    
    .slo-card-link {
        display: inline-block;
        margin-top: 10px;
        font-size: 0.8em;
        color: #667eea;
        text-decoration: none;
    }
    
    .slo-card-link:hover {
        text-decoration: underline;
    }
    
    .no-slos {
        color: #6c757d;
        text-align: center;
        padding: 40px;
    }
    
    .error-message {
        color: #dc3545;
        text-align: center;
        padding: 20px;
    }
    
    .success-details {
        margin-top: 5px;
        font-size: 0.9em;
        opacity: 0.9;
    }
    
    .fa-spinner {
        animation: spin 1s linear infinite;
        margin-right: 8px;
    }
    
    @keyframes spin {
        0% { transform: rotate(0deg); }
        100% { transform: rotate(360deg); }
    }
    
    .template-card {
        cursor: pointer;
        transition: all 0.3s ease;
        animation: fadeIn 0.3s ease;
    }
    
    @keyframes fadeIn {
        from { opacity: 0; transform: translateY(10px); }
        to { opacity: 1; transform: translateY(0); }
    }
    
    .badge {
        margin-right: 5px;
        margin-bottom: 5px;
    }
    
    .badge-primary { background: #e3f2fd; color: #1976d2; }
    .badge-success { background: #e8f5e9; color: #2e7d32; }
    .badge-info { background: #e3f2fd; color: #0288d1; }
    .badge-warning { background: #fff3e0; color: #f57c00; }
    .badge-secondary { background: #f5f5f5; color: #616161; }
    .badge-dark { background: #e0e0e0; color: #212121; }
`;
document.head.appendChild(notificationStyle);

// Initialize
document.addEventListener('DOMContentLoaded', function () {
    console.log('SLOzy initialized with full template support');

    // Fix the Use Template button
    const useTemplateBtn = document.querySelector('button[onclick*="loadTemplates()"]');
    if (useTemplateBtn) {
        useTemplateBtn.onclick = function () {
            showTab('templates');
            loadTemplates();
        };
    }

    setExample('latency');
    loadTemplates();
    loadExistingSLOs();
});

// Global helper
window.showNotification = showNotification;
window.createDefaultTemplate = createDefaultTemplate;