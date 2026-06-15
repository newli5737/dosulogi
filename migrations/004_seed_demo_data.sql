-- Demo seed data (idempotent). Password for demo users: Admin@123

DO $seed$
BEGIN
  IF EXISTS (SELECT 1 FROM customers WHERE code = 'KH-DEMO-01') THEN
    RETURN;
  END IF;

  -- Demo users
  INSERT INTO users (email, password, full_name, role) VALUES
    ('sales1@dosulogi.com', '$2a$10$qHRryOHURzxNn8k7biC6VOvcjiiCXoMiR/QWMR/b67uZSvqerFYrC', 'Nguyễn Minh Tuấn', 'sales_rep'),
    ('sales2@dosulogi.com', '$2a$10$qHRryOHURzxNn8k7biC6VOvcjiiCXoMiR/QWMR/b67uZSvqerFYrC', 'Trần Thị Lan', 'sales_rep'),
    ('ketoan@dosulogi.com', '$2a$10$qHRryOHURzxNn8k7biC6VOvcjiiCXoMiR/QWMR/b67uZSvqerFYrC', 'Lê Văn Hùng', 'accountant'),
    ('giamdoc@dosulogi.com', '$2a$10$qHRryOHURzxNn8k7biC6VOvcjiiCXoMiR/QWMR/b67uZSvqerFYrC', 'Phạm Quốc Bảo', 'director')
  ON CONFLICT (email) DO NOTHING;

  -- Customers
  INSERT INTO customers (code, name, type, email, phone, address, province, segment, tier, tax_code, assigned_to, is_active)
  SELECT v.code, v.name, v.type, v.email, v.phone, v.address, v.province, v.segment, v.tier, v.tax_code,
         (SELECT id FROM users WHERE email = v.assignee LIMIT 1), true
  FROM (VALUES
    ('KH-DEMO-01', 'Công ty TNHH Thương mại Việt Express', 'B2B', 'contact@vietexpress.vn', '02838234567', '12 Nguyễn Huệ, Q.1', 'TP.HCM', 'enterprise', 'gold', '0312345678', 'sales1@dosulogi.com'),
    ('KH-DEMO-02', 'Công ty CP Logistics Đông Nam', 'B2B', 'ops@dongnamlogi.vn', '02837778899', '45 Lê Lợi, Q.1', 'TP.HCM', 'enterprise', 'gold', '0309876543', 'sales1@dosulogi.com'),
    ('KH-DEMO-03', 'Công ty TNHH XNK Minh Phát', 'B2B', 'sales@minhphat.vn', '02435678901', '88 Trần Hưng Đạo, Hoàn Kiếm', 'Hà Nội', 'standard', 'silver', '0101122334', 'sales2@dosulogi.com'),
    ('KH-DEMO-04', 'Shop Online Beauty House', 'B2C', 'hello@beautyhouse.vn', '0909123456', '22 Cách Mạng Tháng 8', 'Đà Nẵng', 'standard', 'standard', NULL, 'sales2@dosulogi.com'),
    ('KH-DEMO-05', 'Công ty TNHH Sản xuất Bao bì Việt', 'B2B', 'info@baobiviet.vn', '02363889900', 'KCN VSIP, Thuận An', 'Bình Dương', 'standard', 'silver', '3700123456', 'sales1@dosulogi.com'),
    ('KH-DEMO-06', 'Công ty CP Thực phẩm Sạch An', 'B2B', 'logistics@ansfood.vn', '02839998877', 'KCN Hiệp Phước, Nhà Bè', 'TP.HCM', 'enterprise', 'gold', '0315566778', 'sales2@dosulogi.com'),
    ('KH-DEMO-07', 'Nguyễn Văn Nam (cá nhân)', 'B2C', 'nam.nguyen@gmail.com', '0918765432', '15 Lê Văn Sỹ, Q.3', 'TP.HCM', 'standard', 'standard', NULL, 'sales1@dosulogi.com'),
    ('KH-DEMO-08', 'Công ty TNHH Dược phẩm Phương Đông', 'B2B', 'supply@phuongdongpharma.vn', '02437776655', 'KCN Sài Đồng, Long Biên', 'Hà Nội', 'enterprise', 'gold', '0108765432', 'sales2@dosulogi.com')
  ) AS v(code, name, type, email, phone, address, province, segment, tier, tax_code, assignee);

  -- Contacts
  INSERT INTO contacts (customer_id, name, role, phone, email, is_primary)
  SELECT c.id, v.cname, v.crole, v.cphone, v.cemail, v.is_primary
  FROM (VALUES
    ('KH-DEMO-01', 'Hoàng Thị Mai', 'Giám đốc điều hành', '0903111222', 'mai.hoang@vietexpress.vn', true),
    ('KH-DEMO-01', 'Vũ Đức Anh', 'Trưởng phòng kho', '0903222333', 'anh.vu@vietexpress.vn', false),
    ('KH-DEMO-02', 'Lý Quốc Huy', 'Trưởng phòng logistics', '0909333444', 'huy.ly@dongnamlogi.vn', true),
    ('KH-DEMO-03', 'Đặng Minh Châu', 'Phụ trách mua hàng', '0912444555', 'chau.dang@minhphat.vn', true),
    ('KH-DEMO-04', 'Phạm Ngọc Linh', 'Chủ shop', '0909555666', 'linh@beautyhouse.vn', true),
    ('KH-DEMO-05', 'Bùi Văn Tài', 'Kế toán trưởng', '0913666777', 'tai.bui@baobiviet.vn', true),
    ('KH-DEMO-06', 'Ngô Thị Hương', 'Supply chain manager', '0908777888', 'huong.ngo@ansfood.vn', true)
  ) AS v(ccode, cname, crole, cphone, cemail, is_primary)
  JOIN customers c ON c.code = v.ccode;

  -- Interactions
  INSERT INTO interactions (customer_id, channel, direction, summary, occurred_at, created_by)
  SELECT c.id, v.channel, v.direction, v.summary, v.occurred_at::timestamptz,
         (SELECT id FROM users WHERE email = 'sales1@dosulogi.com' LIMIT 1)
  FROM (VALUES
    ('KH-DEMO-01', 'call', 'outbound', 'Gọi xác nhận báo giá vận chuyển HCM-HN tháng 6', '2026-06-01 09:30:00'),
    ('KH-DEMO-01', 'email', 'outbound', 'Gửi báo giá dịch vụ fulfillment kho lạnh', '2026-06-03 14:00:00'),
    ('KH-DEMO-02', 'meeting', 'inbound', 'Họp review KPI giao hàng Q2, cam kết OTIF 98%', '2026-06-05 10:00:00'),
    ('KH-DEMO-03', 'call', 'inbound', 'Khách hỏi phí vận chuyển cont 20ft Hải Phòng - HCM', '2026-06-08 16:20:00'),
    ('KH-DEMO-04', 'chat', 'inbound', 'Hỏi tích hợp API tracking cho Shopify', '2026-06-10 11:15:00'),
    ('KH-DEMO-06', 'visit', 'outbound', 'Khảo sát kho tại Hiệp Phước, đề xuất tuyến giao nội thành', '2026-06-12 08:00:00')
  ) AS v(ccode, channel, direction, summary, occurred_at)
  JOIN customers c ON c.code = v.ccode;

  UPDATE customers SET last_contact_at = now() - interval '1 day' * (random() * 14)::int
  WHERE code LIKE 'KH-DEMO-%';

  -- Opportunities
  INSERT INTO opportunities (code, customer_id, title, stage, value, currency, expected_close, assigned_to, created_by, note)
  SELECT v.code, c.id, v.title, v.stage, v.value, 'VND', v.expected_close::date,
         u.id, u.id, v.note
  FROM (VALUES
    ('OPP-DEMO-01', 'KH-DEMO-01', 'Hợp đồng vận chuyển nội địa 2026', 'negotiation', 850000000, '2026-06-30', 'Hợp đồng 12 tháng, 500 chuyến/tháng'),
    ('OPP-DEMO-02', 'KH-DEMO-02', 'Outsourcing kho vận Bình Dương', 'proposal', 420000000, '2026-07-15', 'Cần báo giá 3PL'),
    ('OPP-DEMO-03', 'KH-DEMO-03', 'Vận chuyển XNK đường biển', 'qualified', 310000000, '2026-08-01', 'Tuyến Hải Phòng - Singapore'),
    ('OPP-DEMO-04', 'KH-DEMO-04', 'Gói giao hàng last-mile Đà Nẵng', 'lead', 45000000, '2026-06-25', 'Shop online ~200 đơn/ngày'),
    ('OPP-DEMO-05', 'KH-DEMO-06', 'Cold chain distribution miền Nam', 'won', 1200000000, '2026-05-01', 'Đã chốt, chờ ký HĐ'),
    ('OPP-DEMO-06', 'KH-DEMO-08', 'Phân phối dược phẩm GMP', 'lost', 280000000, '2026-04-20', 'Thua giá đối thủ')
  ) AS v(code, ccode, title, stage, value, expected_close, note)
  JOIN customers c ON c.code = v.ccode
  JOIN users u ON u.email = 'sales1@dosulogi.com';

  INSERT INTO opportunity_stage_history (opportunity_id, from_stage, to_stage, changed_by)
  SELECT o.id, 'proposal', 'negotiation', u.id FROM opportunities o, users u
  WHERE o.code = 'OPP-DEMO-01' AND u.email = 'sales1@dosulogi.com';

  -- Quotations
  INSERT INTO quotations (code, customer_id, opp_id, opportunity_id, items, subtotal, discount, tax_rate, tax_amount, total, currency, valid_until, status, created_by)
  SELECT v.code, c.id, o.id, o.id,
    v.items::jsonb, v.subtotal, v.discount, 10, (v.subtotal - v.discount) * 0.1,
    (v.subtotal - v.discount) * 1.1, 'VND', v.valid_until::date, v.status,
    (SELECT id FROM users WHERE email = 'sales1@dosulogi.com' LIMIT 1)
  FROM (VALUES
    ('BG-DEMO-01', 'KH-DEMO-01', 'OPP-DEMO-01', '[{"description":"Vận chuyển nội địa B2B (500 chuyến/tháng)","qty":12,"unit_price":65000000,"amount":780000000}]', 780000000, 30000000, '2026-07-15', 'sent'),
    ('BG-DEMO-02', 'KH-DEMO-02', 'OPP-DEMO-02', '[{"description":"Thuê kho 3PL 2000m2","qty":12,"unit_price":35000000,"amount":420000000}]', 420000000, 0, '2026-07-30', 'draft'),
    ('BG-DEMO-03', 'KH-DEMO-06', 'OPP-DEMO-05', '[{"description":"Cold chain distribution","qty":12,"unit_price":95000000,"amount":1140000000}]', 1140000000, 50000000, '2026-06-30', 'accepted')
  ) AS v(code, ccode, oppcode, items, subtotal, discount, valid_until, status)
  JOIN customers c ON c.code = v.ccode
  JOIN opportunities o ON o.code = v.oppcode;

  -- Contracts
  INSERT INTO contracts (code, customer_id, opportunity_id, title, start_date, end_date, service_type, value, currency, status, payment_terms, created_by)
  SELECT v.code, c.id, o.id, v.title, v.start_date::date, v.end_date::date, v.service_type, v.value, 'VND', v.status, v.payment_terms,
         (SELECT id FROM users WHERE email = 'sales1@dosulogi.com' LIMIT 1)
  FROM (VALUES
    ('HD-DEMO-01', 'KH-DEMO-06', 'OPP-DEMO-05', 'Hợp đồng cold chain 2026', '2026-05-01', '2027-04-30', 'cold_chain', 1140000000, 'active', 'Net 30'),
    ('HD-DEMO-02', 'KH-DEMO-01', 'OPP-DEMO-01', 'Hợp đồng vận chuyển nội địa', '2026-06-01', '2027-05-31', 'domestic', 825000000, 'draft', 'Net 15')
  ) AS v(code, ccode, oppcode, title, start_date, end_date, service_type, value, status, payment_terms)
  JOIN customers c ON c.code = v.ccode
  LEFT JOIN opportunities o ON o.code = v.oppcode;

  -- Shipments
  INSERT INTO shipments (tracking_code, customer_id, contract_id, status, origin, destination, lat, lng, estimated_delivery, last_synced_at)
  SELECT v.tracking, c.id, ct.id, v.status, v.origin, v.destination, v.lat, v.lng, v.eta::date, now() - interval '2 hours'
  FROM (VALUES
    ('DLX240601001', 'KH-DEMO-01', 'HD-DEMO-02', 'in_transit', 'Kho Q.7, TP.HCM', 'Kho Long Biên, Hà Nội', 15.8500, 108.2000, '2026-06-17'),
    ('DLX240601002', 'KH-DEMO-06', 'HD-DEMO-01', 'delivered', 'KCN Hiệp Phước', 'Siêu thị Coopmart Q.10', 10.6950, 106.7040, '2026-06-14'),
    ('DLX240601003', 'KH-DEMO-04', NULL, 'picked_up', 'Kho Đà Nẵng', 'Khách hàng Sơn Trà', 16.0540, 108.2020, '2026-06-16'),
    ('DLX240601004', 'KH-DEMO-02', NULL, 'pending', 'Kho Bình Dương', 'Kho Cần Thơ', 10.9800, 106.6500, '2026-06-18'),
    ('DLX240601005', 'KH-DEMO-03', NULL, 'in_transit', 'Cảng Hải Phòng', 'Kho Bình Tân', 20.8440, 106.6880, '2026-06-19')
  ) AS v(tracking, ccode, hcode, status, origin, destination, lat, lng, eta)
  JOIN customers c ON c.code = v.ccode
  LEFT JOIN contracts ct ON ct.code = v.hcode;

  INSERT INTO shipment_events (shipment_id, status, description, location, event_time)
  SELECT s.id, v.status, v.description, v.location, v.event_time::timestamptz
  FROM (VALUES
    ('DLX240601001', 'picked_up', 'Lấy hàng tại kho nguồn', 'TP.HCM', '2026-06-15 08:00:00'),
    ('DLX240601001', 'in_transit', 'Xe đang trên cao tốc CT01', 'Đồng Nai', '2026-06-15 14:30:00'),
    ('DLX240601002', 'delivered', 'Giao hàng thành công', 'TP.HCM', '2026-06-14 16:45:00'),
    ('DLX240601003', 'picked_up', 'Shipper đã lấy hàng', 'Đà Nẵng', '2026-06-15 10:00:00')
  ) AS v(tracking, status, description, location, event_time)
  JOIN shipments s ON s.tracking_code = v.tracking;

  INSERT INTO opportunity_shipments (opportunity_id, shipment_id)
  SELECT o.id, s.id FROM opportunities o, shipments s
  WHERE o.code = 'OPP-DEMO-01' AND s.tracking_code = 'DLX240601001';

  -- Tickets
  INSERT INTO tickets (code, customer_id, title, description, priority, status, category, assigned_to, sla_deadline, created_by)
  SELECT v.code, c.id, v.title, v.description, v.priority, v.status, v.category,
         (SELECT id FROM users WHERE email = v.assignee LIMIT 1),
         (now() + v.sla_hours * interval '1 hour'),
         (SELECT id FROM users WHERE email = 'sales1@dosulogi.com' LIMIT 1)
  FROM (VALUES
    ('TK-DEMO-01', 'KH-DEMO-01', 'Chậm giao lô hàng HCM-HN', 'Khách phản ánh trễ 4h so với ETA', 'high', 'in_progress', 'delivery', 'sales1@dosulogi.com', 8),
    ('TK-DEMO-02', 'KH-DEMO-04', 'Hỏng hàng khi giao', '2 thùng mỹ phẩm bị móp', 'urgent', 'open', 'claim', 'sales2@dosulogi.com', 4),
    ('TK-DEMO-03', 'KH-DEMO-06', 'Yêu cầu báo cáo nhiệt độ cold chain', 'Cần log nhiệt độ 7 ngày gần nhất', 'medium', 'open', 'report', 'sales1@dosulogi.com', 24),
    ('TK-DEMO-04', 'KH-DEMO-02', 'Cập nhật địa chỉ giao mới', 'Đổi điểm giao sang KCN Mỹ Phước', 'low', 'resolved', 'address', 'sales1@dosulogi.com', 72)
  ) AS v(code, ccode, title, description, priority, status, category, assignee, sla_hours)
  JOIN customers c ON c.code = v.ccode;

  INSERT INTO ticket_comments (ticket_id, body, is_internal, created_by)
  SELECT t.id, v.body, v.internal, u.id
  FROM (VALUES
    ('TK-DEMO-01', 'Đã liên hệ tài xế, xe dự kiến đến HN 20h tối nay', true, 'sales1@dosulogi.com'),
    ('TK-DEMO-01', 'Cập nhật cho khách qua email', false, 'sales1@dosulogi.com'),
    ('TK-DEMO-02', 'Ticket được tạo tự động từ hệ thống', true, 'sales2@dosulogi.com')
  ) AS v(tcode, body, internal, uemail)
  JOIN tickets t ON t.code = v.tcode
  JOIN users u ON u.email = v.uemail;

  -- Invoices
  INSERT INTO invoices (code, customer_id, contract_id, items, subtotal, tax_rate, tax_amount, total, currency, status, due_date, created_by)
  SELECT v.code, c.id, ct.id,
    v.items::jsonb, v.subtotal, 10, v.subtotal * 0.1, v.subtotal * 1.1, 'VND', v.status, v.due_date::date,
    (SELECT id FROM users WHERE email = 'ketoan@dosulogi.com' LIMIT 1)
  FROM (VALUES
    ('HD-DEMO-INV-01', 'KH-DEMO-06', 'HD-DEMO-01', '[{"description":"Cold chain T5/2026","qty":1,"unit_price":95000000,"amount":95000000}]', 95000000, 'sent', '2026-06-30'),
    ('HD-DEMO-INV-02', 'KH-DEMO-01', 'HD-DEMO-02', '[{"description":"Vận chuyển nội địa T6/2026","qty":1,"unit_price":68750000,"amount":68750000}]', 68750000, 'draft', '2026-07-15'),
    ('HD-DEMO-INV-03', 'KH-DEMO-02', NULL, '[{"description":"Thuê kho tháng 5","qty":1,"unit_price":35000000,"amount":35000000}]', 35000000, 'overdue', '2026-05-31')
  ) AS v(code, ccode, hcode, items, subtotal, status, due_date)
  JOIN customers c ON c.code = v.ccode
  LEFT JOIN contracts ct ON ct.code = v.hcode;

  INSERT INTO payments (invoice_id, amount, method, reference_code, matched_auto, note)
  SELECT i.id, 95000000, 'bank_transfer', 'SEPAY-DEMO-001', true, 'Thanh toán đủ cold chain T5'
  FROM invoices i WHERE i.code = 'HD-DEMO-INV-01';

  UPDATE invoices SET status = 'paid', paid_at = now() - interval '3 days' WHERE code = 'HD-DEMO-INV-01';

  -- Campaigns
  INSERT INTO campaigns (name, type, status, subject, body_html, sent_count, created_by)
  SELECT v.name, v.type, v.status, v.subject, v.body_html, v.sent_count,
         (SELECT id FROM users WHERE email = 'admin@dosulogi.com' LIMIT 1)
  FROM (VALUES
    ('Newsletter Q2 2026', 'email', 'sent', 'Dosu Logi — Cập nhật dịch vụ cold chain', '<p>Kính gửi quý khách...</p>', 128),
    ('Khuyến mãi last-mile Đà Nẵng', 'email', 'draft', 'Giảm 10% phí giao nội thành', '<p>Ưu đãi tháng 6...</p>', 0),
    ('Nhắc công nợ tháng 5', 'email', 'scheduled', 'Thông báo công nợ đến hạn', '<p>Quý khách vui lòng thanh toán...</p>', 0)
  ) AS v(name, type, status, subject, body_html, sent_count);

  UPDATE campaigns SET scheduled_at = now() + interval '2 days' WHERE name = 'Nhắc công nợ tháng 5';

END $seed$;
