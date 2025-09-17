    function centerOnUser() {
            const lat = parseFloat(document.getElementById('lat').value);
            const lng = parseFloat(document.getElementById('lng').value);
            
            if (lat && lng && map) {
                map.setView([lat, lng], 16);
                if (userMarker) {
                    userMarker.openPopup();
                }
            } else {
                getLocation();
            }
        }
        
        // Toggle pantalla completa para el mapa
        function toggleFullscreen() {
            const mapContainer = document.querySelector('.map-container');
            
            if (!document.fullscreenElement) {
                mapContainer.requestFullscreen().then(() => {
                    mapContainer.style.height = '100vh';
                    setTimeout(() => map.invalidateSize(), 100);
                });
            } else {
                document.exitFullscreen().then(() => {
                    mapContainer.style.height = '500px';
                    setTimeout(() => map.invalidateSize(), 100);
                });
            }
        }
        
        // Manejo de eventos de pantalla completa
        document.addEventListener('fullscreenchange', function() {
            if (map) {
                setTimeout(() => map.invalidateSize(), 200);
            }
        });
        
        // Contador de caracteres para descripción
        document.addEventListener('DOMContentLoaded', function() {
            const descField = document.getElementById('description');
            if (descField) {
                const helpText = document.getElementById('desc-help');
                descField.addEventListener('input', function() {
                    const remaining = 200 - this.value.length;
                    helpText.textContent = `${remaining} caracteres restantes`;
                    
                    if (remaining < 20) {
                        helpText.style.color = '#f44336';
                    } else {
                        helpText.style.color = '#666';
                    }
                });
            }
        });
        
        // Precargar ubicación si hay permisos
        if ('geolocation' in navigator) {
            navigator.permissions.query({name: 'geolocation'}).then(function(result) {
                if (result.state === 'granted') {
                    // Obtener ubicación silenciosamente al cargar
                    navigator.geolocation.getCurrentPosition(function(position) {
                        // No actualizar automáticamente, solo precargar
                        console.log('📍 Ubicación precargada:', position.coords.latitude, position.coords.longitude);
                    }, function() {
                        // Ignorar errores en precarga
                    });
                }
            });
        }