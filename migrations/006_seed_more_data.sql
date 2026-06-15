-- Additional demo seed (idempotent)

DO $seed$
BEGIN
  IF EXISTS (SELECT 1 FROM opportunities WHERE code = 'OPP-DEMO-07') THEN
    RETURN;
  END IF;

  INSERT INTO customers (code, name, type, email, phone, address, province, segment, tier, tax_code, assigned_to, is_active)
  SELECT v.code, v.name, v.type, v.email, v.phone, v.address, v.province, v.segment, v.tier, v.tax_code,
         (SELECT id FROM users WHERE email = v.assignee LIMIT 1), true
  FROM (VALUES
    ('KH-DEMO-09', 'Công ty CP Thương mại điện tử Nova', 'B2B', 'ops@novacom.vn', '02835556677', 'KCN Tân Bình, TP.HCM', 'TP.HCM', 'enterprise', 'gold', '0319988776', 'sales2@dosulogi.com'),
    ('KH-DEMO-10', 'Công ty TNHH Nông sản Xanh Việt', 'B2B', 'logi@xanhviet.vn', '02438889900', 'KCN Bắc Thăng Long, Đông Anh', 'Hà Nội', 'standard', 'silver', '0105566778', 'sales1@dosulogi.com'),
    ('KH-DEMO-11', 'Công ty CP Dược Hậu Giang', 'B2B', 'supply@dhg.vn', '02923887766', 'KCN Mỹ Tho, Tiền Giang', 'Tiền Giang', 'enterprise', 'gold', '1200123456', 'sales2@dosulogi.com'),
    ('KH-DEMO-12', 'FMCG Miền Trung JSC', 'B2B', 'warehouse@fmcgmt.vn', '02363881122', 'KCN Hòa Khánh, Đà Nẵng', 'Đà Nẵng', 'standard', 'silver', '0400123987', 'sales1@dosulogi.com')
  ) AS v(code, name, type, email, phone, address, province, segment, tier, tax_code, assignee);

  INSERT INTO opportunities (code, customer_id, title, stage, value, currency, expected_close, assigned_to, created_by, note)
  SELECT v.code, c.id, v.title, v.stage, v.value, 'VND', v.expected_close::date, u.id, u.id, v.note
  FROM (VALUES
    ('OPP-DEMO-07', 'KH-DEMO-09', 'Fulfillment đa kênh Shopee/Lazada', 'proposal', 560000000, '2026-07-20', 'Tích hợp API vận đơn'),
    ('OPP-DEMO-08', 'KH-DEMO-10', 'Vận chuyển nông sản tươi Bắc-Nam', 'qualified', 190000000, '2026-08-10', 'Xe lạnh 5 tấn'),
    ('OPP-DEMO-09', 'KH-DEMO-11', 'Phân phối thuốc GMP miền Nam', 'negotiation', 720000000, '2026-06-28', 'Đang thương thảo SLA'),
    ('OPP-DEMO-10', 'KH-DEMO-12', 'Kho bãi + giao nội thành Đà Nẵng', 'lead', 88000000, '2026-07-05', '200 đơn/ngày'),
    ('OPP-DEMO-11', 'KH-DEMO-03', 'Cont 40RF Hải Phòng — Singapore', 'won', 450000000, '2026-05-15', 'Đã ký HĐ sea freight'),
    ('OPP-DEMO-12', 'KH-DEMO-05', 'Thuê kho Bình Dương 5000m2', 'lost', 320000000, '2026-04-30', 'Khách chọn đối thủ giá rẻ')
  ) AS v(code, ccode, title, stage, value, expected_close, note)
  JOIN customers c ON c.code = v.ccode
  JOIN users u ON u.email = 'sales1@dosulogi.com';

  INSERT INTO quotations (code, customer_id, opp_id, opportunity_id, items, subtotal, discount, tax_rate, tax_amount, total, currency, valid_until, status, created_by)
  SELECT v.code, c.id, o.id, o.id, v.items::jsonb, v.subtotal, v.discount, 10, (v.subtotal - v.discount) * 0.1,
    (v.subtotal - v.discount) * 1.1, 'VND', v.valid_until::date, v.status,
    (SELECT id FROM users WHERE email = 'sales1@dosulogi.com' LIMIT 1)
  FROM (VALUES
    ('BG-DEMO-04', 'KH-DEMO-09', 'OPP-DEMO-07', '[{"description":"Fulfillment đa kênh","qty":12,"unit_price":45000000,"amount":540000000}]', 540000000, 20000000, '2026-08-01', 'sent'),
    ('BG-DEMO-05', 'KH-DEMO-11', 'OPP-DEMO-09', '[{"description":"Phân phối GMP","qty":12,"unit_price":58000000,"amount":696000000}]', 696000000, 0, '2026-07-10', 'sent'),
    ('BG-DEMO-06', 'KH-DEMO-03', 'OPP-DEMO-11', '[{"description":"Sea freight cont lạnh","qty":24,"unit_price":18000000,"amount":432000000}]', 432000000, 0, '2026-06-15', 'accepted')
  ) AS v(code, ccode, oppcode, items, subtotal, discount, valid_until, status)
  JOIN customers c ON c.code = v.ccode
  JOIN opportunities o ON o.code = v.oppcode;

  INSERT INTO contracts (code, customer_id, opportunity_id, title, start_date, end_date, service_type, value, currency, status, payment_terms, created_by)
  SELECT v.code, c.id, o.id, v.title, v.start_date::date, v.end_date::date, v.service_type, v.value, 'VND', v.status, v.payment_terms,
         (SELECT id FROM users WHERE email = 'sales1@dosulogi.com' LIMIT 1)
  FROM (VALUES
    ('HD-DEMO-03', 'KH-DEMO-03', 'OPP-DEMO-11', 'Hợp đồng vận tải biển lạnh', '2026-05-15', '2027-05-14', 'sea', 432000000, 'active', 'Net 30'),
    ('HD-DEMO-04', 'KH-DEMO-09', 'OPP-DEMO-07', 'Hợp đồng fulfillment TMĐT', '2026-07-01', '2027-06-30', 'warehouse', 540000000, 'draft', 'Net 15'),
    ('HD-DEMO-05', 'KH-DEMO-12', 'OPP-DEMO-10', 'Last-mile Đà Nẵng', '2026-07-01', '2026-12-31', 'last_mile', 88000000, 'draft', 'Net 7')
  ) AS v(code, ccode, oppcode, title, start_date, end_date, service_type, value, status, payment_terms)
  JOIN customers c ON c.code = v.ccode
  LEFT JOIN opportunities o ON o.code = v.oppcode;

  INSERT INTO shipments (tracking_code, customer_id, contract_id, status, origin, destination, lat, lng, estimated_delivery, last_synced_at)
  SELECT v.tracking, c.id, ct.id, v.status, v.origin, v.destination, v.lat, v.lng, v.eta::date, now() - interval '1 hour'
  FROM (VALUES
    ('DLX240602001', 'KH-DEMO-09', 'HD-DEMO-04', 'in_transit', 'Kho Tân Bình', 'Hub Giao Hàng Nhanh Q.12', 10.8010, 106.6520, '2026-06-16'),
    ('DLX240602002', 'KH-DEMO-10', NULL, 'picked_up', 'KCN Bắc Thăng Long', 'Siêu thị BigC Long Biên', 21.0500, 105.8500, '2026-06-16'),
    ('DLX240602003', 'KH-DEMO-11', NULL, 'pending', 'Kho DHG Mỹ Tho', 'Nhà thuốc Long An', 10.3600, 106.3600, '2026-06-17'),
    ('DLX240602004', 'KH-DEMO-12', 'HD-DEMO-05', 'out_for_delivery', 'Hub Đà Nẵng', 'Khách Sơn Trà', 16.0700, 108.2200, '2026-06-15'),
    ('DLX240602005', 'KH-DEMO-03', 'HD-DEMO-03', 'in_transit', 'Cảng Hải Phòng', 'Cảng Singapore', 20.8440, 106.6880, '2026-06-22'),
    ('DLX240602006', 'KH-DEMO-06', 'HD-DEMO-01', 'delivered', 'KCN Hiệp Phước', 'WinMart Q.7', 10.6950, 106.7040, '2026-06-14')
  ) AS v(tracking, ccode, hcode, status, origin, destination, lat, lng, eta)
  JOIN customers c ON c.code = v.ccode
  LEFT JOIN contracts ct ON ct.code = v.hcode;

  INSERT INTO shipment_events (shipment_id, status, description, location, event_time)
  SELECT s.id, v.status, v.description, v.location, v.event_time::timestamptz
  FROM (VALUES
    ('DLX240602001', 'picked_up', 'Nhận hàng tại kho', 'TP.HCM', '2026-06-15 07:00:00'),
    ('DLX240602001', 'in_transit', 'Đang phân loại hub', 'TP.HCM', '2026-06-15 11:30:00'),
    ('DLX240602004', 'out_for_delivery', 'Shipper đang giao', 'Đà Nẵng', '2026-06-15 15:00:00'),
    ('DLX240602005', 'in_transit', 'Cont đã lên tàu', 'Hải Phòng', '2026-06-14 20:00:00')
  ) AS v(tracking, status, description, location, event_time)
  JOIN shipments s ON s.tracking_code = v.tracking;

  INSERT INTO tickets (code, customer_id, title, description, priority, status, category, assigned_to, sla_deadline, created_by)
  SELECT v.code, c.id, v.title, v.description, v.priority, v.status, v.category,
         (SELECT id FROM users WHERE email = v.assignee LIMIT 1),
         (now() + v.sla_hours * interval '1 hour'),
         (SELECT id FROM users WHERE email = 'sales1@dosulogi.com' LIMIT 1)
  FROM (VALUES
    ('TK-DEMO-05', 'KH-DEMO-09', 'Sai mã vận đơn TMĐT', '50 đơn Shopee sync sai tracking', 'high', 'open', 'integration', 'sales2@dosulogi.com', 12),
    ('TK-DEMO-06', 'KH-DEMO-10', 'Nhiệt độ xe lạnh cao', 'Cảm biến báo 8°C vượt ngưỡng', 'urgent', 'in_progress', 'cold_chain', 'sales1@dosulogi.com', 4),
    ('TK-DEMO-07', 'KH-DEMO-11', 'Yêu cầu COA lô hàng', 'Cần chứng nhận nhiệt độ giao hàng', 'medium', 'open', 'report', 'sales2@dosulogi.com', 48),
    ('TK-DEMO-08', 'KH-DEMO-12', 'Tăng volume giao cuối tuần', 'Xin bổ sung 2 shipper T7-CN', 'low', 'pending', 'capacity', 'sales1@dosulogi.com', 72)
  ) AS v(code, ccode, title, description, priority, status, category, assignee, sla_hours)
  JOIN customers c ON c.code = v.ccode;

  INSERT INTO invoices (code, customer_id, contract_id, items, subtotal, tax_rate, tax_amount, total, currency, status, due_date, created_by)
  SELECT v.code, c.id, ct.id, v.items::jsonb, v.subtotal, 10, v.subtotal * 0.1, v.subtotal * 1.1, 'VND', v.status, v.due_date::date,
    (SELECT id FROM users WHERE email = 'ketoan@dosulogi.com' LIMIT 1)
  FROM (VALUES
    ('HD-DEMO-INV-04', 'KH-DEMO-03', 'HD-DEMO-03', '[{"description":"Sea freight T6/2026","qty":1,"unit_price":36000000,"amount":36000000}]', 36000000, 'sent', '2026-07-10'),
    ('HD-DEMO-INV-05', 'KH-DEMO-09', NULL, '[{"description":"Fulfillment pilot T6","qty":1,"unit_price":42000000,"amount":42000000}]', 42000000, 'draft', '2026-07-20'),
    ('HD-DEMO-INV-06', 'KH-DEMO-06', 'HD-DEMO-01', '[{"description":"Cold chain T6/2026","qty":1,"unit_price":95000000,"amount":95000000}]', 95000000, 'sent', '2026-07-05')
  ) AS v(code, ccode, hcode, items, subtotal, status, due_date)
  JOIN customers c ON c.code = v.ccode
  LEFT JOIN contracts ct ON ct.code = v.hcode;

  INSERT INTO campaigns (name, type, status, subject, body_html, sent_count, created_by)
  SELECT v.name, v.type, v.status, v.subject, v.body_html, v.sent_count,
         (SELECT id FROM users WHERE email = 'admin@dosulogi.com' LIMIT 1)
  FROM (VALUES
    ('Giới thiệu dịch vụ sea freight', 'email', 'sent', 'Dosu Logi — Vận tải biển & cold chain', '<p>Giải pháp logistics toàn diện...</p>', 86),
    ('Webinar ERP logistics', 'email', 'scheduled', 'Mời tham dự webinar vận hành kho', '<p>Đăng ký miễn phí...</p>', 0)
  ) AS v(name, type, status, subject, body_html, sent_count);

END $seed$;
