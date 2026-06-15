import { useEffect, useState } from 'react'
import { MapContainer, Marker, Popup, TileLayer } from 'react-leaflet'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { getShipmentMap } from '../api'
import { Page, useToken } from '../components/ui'

import markerIcon2x from 'leaflet/dist/images/marker-icon-2x.png'
import markerIcon from 'leaflet/dist/images/marker-icon.png'
import markerShadow from 'leaflet/dist/images/marker-shadow.png'

delete L.Icon.Default.prototype._getIconUrl
L.Icon.Default.mergeOptions({ iconRetinaUrl: markerIcon2x, iconUrl: markerIcon, shadowUrl: markerShadow })

export default function ShipmentMapPage() {
  const token = useToken()
  const [points, setPoints] = useState([])

  useEffect(() => {
    getShipmentMap(token).then(setPoints).catch(console.error)
  }, [token])

  const center = points[0] ? [points[0].lat, points[0].lng] : [16.0544, 108.2022]

  return (
    <Page title="Bản đồ vận đơn (Leaflet / OSM)">
      <div className="map-box">
        <MapContainer center={center} zoom={6} scrollWheelZoom style={{ height: '100%', width: '100%' }}>
          <TileLayer attribution='&copy; OpenStreetMap' url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png" />
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
    </Page>
  )
}
