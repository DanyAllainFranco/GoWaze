
// Variables globales
let map;
let userMarker;
let reportMarkers = [];
let routeControl;
let ws;
let reconnectInterval;
let routeWaypoints = [];

// Inicialización cuando carga el DOM
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
});

// Inicializar aplicación
function initializeApp() {
    console.log('🚀 Inicializando GoWaze...');
    initMap();
    connectWebSocket();
    setupEventListeners();
    console.log('✅ GoWaze inicializado correctamente');
}

// Inicializar mapa
function initMap() {
    console.log('🗺️ Inicializando mapa...');
    
    // Crear mapa centrado en San Pedro Sula
    map = L.map('map').setView([14.0818, -87.2068], 13);

    // Agregar capa de OpenStreetMap
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© OpenStreetMap contributors',
        maxZoom: 19
    }).addTo(map);

    // Configurar eventos del mapa
    setupMapEvents();
    
    console.log('✅ Mapa inicializado');
}

// Configurar eventos del mapa
function setupMapEvents() {
    // Click derecho para reportar
    map.on('contextmenu', function(e) {
        const lat = e.latlng.lat.toFixed(6);
        const lng = e.latlng.lng.toFixed(6);
        
        document.getElementById('report-lat').value = lat;
        document.getElementById('report-lng').value = lng;
        
        // Crear popup temporal
        L.popup()
            .setLatLng(e.latlng)
            .setContent(`📍 Ubicación seleccionada para reporte<br>Lat: ${lat}<br>Lng: ${lng}`)
            .openOn(map);
        
        console.log(`📍 Ubicación seleccionada: ${lat}, ${lng}`);
    });

    // Click para agregar waypoints de ruta
    map.on('click', function(e) {
        if (routeWaypoints.length < 2) {
            routeWaypoints.push(e.latlng);
            
            const icon = routeWaypoints.length === 1 ? '🅰️' : '🅱️';
            L.marker(e.latlng, {
                icon: L.divIcon({
                    html: `<div style="background: white; border-radius: 50%; width: 30px; height: 30px; display: flex; align-items: center; justify-content: center; border: 2px solid #4CAF50; font-weight: bold;">${icon}</div>`,
                    iconSize: [30, 30],
                    iconAnchor: [15, 15]
                })
            }).addTo(map);
            
            console.log(`${icon} Waypoint agregado: ${e.latlng.lat}, ${e.latlng.lng}`);
            
            if (routeWaypoints.length === 2) {
                calculateRouteOnMap();
            }
        }
    });
}

// Configurar event listeners
function setupEventListeners() {
    // Enter en búsqueda
    const searchInput = document.getElementById('search-input');
    if (searchInput) {
        searchInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                searchPlace();
            }
        });
    }

    // Evento después de enviar formulario de usuario
    document.body.addEventListener('htmx:afterRequest', function(event) {
        if (event.detail.pathInfo.requestPath === '/api/users') {
            const lat = parseFloat(document.getElementById('lat').value);
            const lng = parseFloat(document.getElementById('lng').value);
            updateUserMarker(lat, lng);
            map.setView([lat, lng], 15);
        }
    });

    console.log('👂 Event listeners configurados');
}

// Buscar lugar usando Nominatim
async function searchPlace() {
    const query = document.getElementById('search-input').value;
    if (!query.trim()) {
        showSearchResult('❌ Ingresa un lugar para buscar', 'error');
        return;
    }

    console.log(`🔍 Buscando: ${query}`);
    showSearchResult('🔍 Buscando...', 'loading');

    try {
        const response = await fetch(`https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(query)}&limit=1`);
        const data = await response.json();
        
        if (data.length > 0) {
            const place = data[0];
            const lat = parseFloat(place.lat);
            const lng = parseFloat(place.lon);
            
            map.setView([lat, lng], 15);
            
            L.popup()
                .setLatLng([lat, lng])
                .setContent(`📍 ${place.display_name}`)
                .openOn(map);
            
            showSearchResult(`✅ Encontrado: ${place.display_name}`, 'success');
            console.log(`✅ Lugar encontrado: ${place.display_name}`);
        } else {
            showSearchResult('❌ Lugar no encontrado', 'error');
            console.log('❌ No se encontró el lugar');
        }
    } catch (error) {
        console.error('Error buscando lugar:', error);
        showSearchResult('❌ Error en la búsqueda', 'error');
    }
}

// Mostrar resultado de búsqueda
function showSearchResult(message, type) {
    const resultsDiv = document.getElementById('search-results');
    if (resultsDiv) {
        const className = type === 'success' ? 'green' : type === 'error' ? 'red' : '#666';
        resultsDiv.innerHTML = `<div style="color: ${className}; margin-top: 10px;">${message}</div>`;
        
        // Limpiar después de 5 segundos si es éxito
        if (type === 'success') {
            setTimeout(() => {
                resultsDiv.innerHTML = '';
            }, 5000);
        }
    }
}

// Obtener ubicación GPS
function getLocation() {
    console.log('🌍 Obteniendo ubicación GPS...');
    
    if (!navigator.geolocation) {
        alert("Geolocalización no es soportada por este navegador.");
        return;
    }

    navigator.geolocation.getCurrentPosition(
        function(position) {
            const lat = position.coords.latitude;
            const lng = position.coords.longitude;
            
            document.getElementById('lat').value = lat.toFixed(6);
            document.getElementById('lng').value = lng.toFixed(6);
            
            map.setView([lat, lng], 15);
            updateUserMarker(lat, lng);
            
            console.log(`✅ Ubicación GPS obtenida: ${lat}, ${lng}`);
        },
        function(error) {
            console.error('Error GPS:', error);
            let errorMsg = 'Error obteniendo ubicación: ';
            switch(error.code) {
                case error.PERMISSION_DENIED:
                    errorMsg += 'Permisos denegados';
                    break;
                case error.POSITION_UNAVAILABLE:
                    errorMsg += 'Ubicación no disponible';
                    break;
                case error.TIMEOUT:
                    errorMsg += 'Tiempo de espera agotado';
                    break;
                default:
                    errorMsg += error.message;
                    break;
            }
            alert(errorMsg);
        },
        {
            enableHighAccuracy: true,
            timeout: 10000,
            maximumAge: 60000
        }
    );
}

// Actualizar marcador de usuario
function updateUserMarker(lat, lng) {
    if (userMarker) {
        map.removeLayer(userMarker);
    }
    
    userMarker = L.marker([lat, lng], {
        icon: L.divIcon({
            html: '<div class="custom-marker marker-user">👤</div>',
            iconSize: [20, 20],
            iconAnchor: [10, 10]
        })
    }).addTo(map);
    
    userMarker.bindPopup('📍 Tu ubicación').openPopup();
    console.log(`👤 Marcador de usuario actualizado: ${lat}, ${lng}`);
}

// Calcular ruta en el mapa
function calculateRoute() {
    if (routeWaypoints.length === 2) {
        calculateRouteOnMap();
    } else {
        alert('Click en 2 puntos del mapa para calcular la ruta (origen y destino)');
        console.log('ℹ️ Se necesitan 2 waypoints para calcular ruta');
    }
}

function calculateRouteOnMap() {
    console.log('🧭 Calculando ruta...');
    
    if (routeControl) {
        map.removeControl(routeControl);
    }

    routeControl = L.Routing.control({
        waypoints: routeWaypoints,
        routeWhileDragging: true,
        addWaypoints: false,
        createMarker: function() { return null; }, // No crear marcadores automáticos
        lineOptions: {
            styles: [{ color: '#4CAF50', weight: 6, opacity: 0.8 }]
        },
        router: L.Routing.osrmv1({
            serviceUrl: 'https://router.project-osrm.org/route/v1'
        })
    }).addTo(map);

    routeControl.on('routesfound', function(e) {
        const route = e.routes[0];
        const distance = (route.summary.totalDistance / 1000).toFixed(2);
        const time = Math.round(route.summary.totalTime / 60);
        
        const routeInfo = `
            <div class="route-info">
                <strong>📊 Información de Ruta:</strong><br>
                📏 Distancia: ${distance} km<br>
                ⏱️ Tiempo: ${time} minutos<br>
                🛣️ Puntos: ${route.coordinates.length}
            </div>
        `;
        
        document.getElementById('route-info').innerHTML = routeInfo;
        console.log(`✅ Ruta calculada: ${distance} km, ${time} min`);
    });

    routeControl.on('routingerror', function(e) {
        console.error('Error calculando ruta:', e);
        document.getElementById('route-info').innerHTML = 
            '<div style="color: red;">❌ Error calculando ruta</div>';
    });
}

// Limpiar ruta
function clearRoute() {
    console.log('🗑️ Limpiando ruta...');
    
    if (routeControl) {
        map.removeControl(routeControl);
        routeControl = null;
    }
    
    routeWaypoints = [];
    document.getElementById('route-info').innerHTML = '';
    
    // Limpiar marcadores de ruta
    map.eachLayer(function(layer) {
        if (layer instanceof L.Marker && layer !== userMarker && !reportMarkers.includes(layer)) {
            map.removeLayer(layer);
        }
    });
    
    console.log('✅ Ruta limpiada');
}

// Actualizar marcadores de reportes
function updateReportMarkers(reports) {
    console.log(`📍 Actualizando ${reports.length} marcadores de reportes`);
    
    // Limpiar marcadores existentes
    reportMarkers.forEach(marker => map.removeLayer(marker));
    reportMarkers = [];

    reports.forEach(report => {
        const icons = {
            accident: '🚗',
            police: '👮',
            traffic: '🚦',
            hazard: '⚠️'
        };

        const marker = L.marker([report.lat, report.lng], {
            icon: L.divIcon({
                html: `<div class="custom-marker marker-${report.type}">${icons[report.type]}</div>`,
                iconSize: [20, 20],
                iconAnchor: [10, 10]
            })
        }).addTo(map);

        marker.bindPopup(`
            <div style="min-width: 200px;">
                <strong>${icons[report.type]} ${report.type.toUpperCase()}</strong><br>
                <p style="margin: 10px 0;">${report.description}</p>
                <small style="color: #666;">
                    📅 ${new Date(report.created_at).toLocaleString()}<br>
                    👍 ${report.votes} votos
                </small>
            </div>
        `);

        reportMarkers.push(marker);
    });
}

// WebSocket - Conectar
function connectWebSocket() {
    console.log('🔌 Conectando WebSocket...');
    
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    ws = new WebSocket(wsUrl);
    
    ws.onopen = function() {
        console.log('✅ WebSocket conectado');
        document.getElementById('status').textContent = 'Conectado';
        document.getElementById('status').className = 'connected';
        clearInterval(reconnectInterval);
        
        // Enviar ping cada 30 segundos para mantener conexión
        setInterval(() => {
            if (ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({ type: 'ping' }));
            }
        }, 30000);
    };
    
    ws.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            handleWebSocketMessage(data);
        } catch (error) {
            console.error('Error procesando mensaje WebSocket:', error);
        }
    };
    
    ws.onclose = function() {
        console.log('❌ WebSocket desconectado');
        document.getElementById('status').textContent = 'Desconectado';
        document.getElementById('status').className = 'disconnected';
        
        // Intentar reconectar cada 3 segundos
        reconnectInterval = setInterval(connectWebSocket, 3000);
    };
    
    ws.onerror = function(error) {
        console.error('Error WebSocket:', error);
    };
}

// Manejar mensajes WebSocket
function handleWebSocketMessage(data) {
    console.log('📨 Mensaje WebSocket recibido:', data.type);
    
    // Actualizar estadísticas
    updateStats(data);
    
    // Actualizar reportes en mapa
    if (data.reports) {
        updateReportMarkers(data.reports);
    }
    
    // Manejar tipos específicos
    switch (data.type) {
        case 'new_report':
            console.log('🚨 Nuevo reporte recibido');
            break;
        case 'stats':
            // Ya manejado en updateStats
            break;
        default:
            console.log('📨 Tipo de mensaje desconocido:', data.type);
    }
}

// Actualizar estadísticas en UI
function updateStats(data) {
    if (data.users_online !== undefined) {
        document.getElementById('users-online').textContent = data.users_online;
    }
    if (data.total_reports !== undefined) {
        document.getElementById('total-reports').textContent = data.total_reports;
    }
    if (data.traffic_points !== undefined) {
        document.getElementById('traffic-points').textContent = data.traffic_points;
    }
}

// Funciones de utilidad
function showNotification(message, type = 'info') {
    console.log(`📢 ${type.toUpperCase()}: ${message}`);
    // Aquí podrías agregar notificaciones toast en el futuro
}

// Manejar errores globales
window.addEventListener('error', function(e) {
    console.error('Error global:', e.error);
});

// Log de carga completa
window.addEventListener('load', function() {
    console.log('🎉 GoWaze cargado completamente');
});