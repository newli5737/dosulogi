import { useEffect, useState } from 'react'
import { MapContainer, Marker, Popup, TileLayer } from 'react-leaflet'
import L, { type LatLngTuple } from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { shipmentApi } from '@/entities/shipment/api/shipmentApi'
import type { MapPoint } from '@/entities/shipment/model/types'
import './shipment-map.css'

import markerIcon2x from 'leaflet/dist/images/marker-icon-2x.png'
import markerIcon from 'leaflet/dist/images/marker-icon.png'
import markerShadow from 'leaflet/dist/images/marker-shadow.png'

delete (L.Icon.Default.prototype as { _getIconUrl?: unknown })._getIconUrl
L.Icon.Default.mergeOptions({ iconRetinaUrl: markerIcon2x, iconUrl: markerIcon, shadowUrl: markerShadow })

const DEFAULT_CENTER: LatLngTuple = [16.0544, 108.2022]

export function ShipmentMapPage() {
  const [points, setPoints] = useState<MapPoint[]>([])

  useEffect(() => {
    shipmentApi.map().then((d) => setPoints(Array.isArray(d) ? d : [])).catch(console.error)
  }, [])

  const center: LatLngTuple = points[0] ? [points[0].lat, points[0].lng] : DEFAULT_CENTER

  return (
    <div className="page-card">
      <div className="page-header"><h1>Bản đồ vận đơn</h1></div>
      <div className="map-box">
        <MapContainer center={center} zoom={6} scrollWheelZoom style={{ height: '100%', width: '100%' }}>
          <TileLayer attribution="&copy; OpenStreetMap" url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png" />
          {points.map((p) => (
            <Marker key={p.tracking_code} position={[p.lat, p.lng]}>
              <Popup>
                <strong>{p.tracking_code}</strong><br />
                {p.status}<br />
                {p.customer_name}<br />
                {p.destination}
              </Popup>
            </Marker>
          ))}
        </MapContainer>
      </div>
    </div>
  )
}
