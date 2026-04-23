-- SuperSOL Demo Seed Data
-- 고객: 김지원

INSERT INTO customers (name, phone, membership) VALUES
('김지원', '010-1234-5678', 'PREMIER');

-- 계좌
INSERT INTO accounts (customer_id, account_number, account_name, bank_name, balance, account_type, is_primary) VALUES
(1, '110-456-789012', '신한 주거래통장',   '신한은행', 11234510, 'CHECKING', TRUE),
(1, '110-789-012345', '신한 급여통장',     '신한은행', 3200000,  'CHECKING', FALSE),
(1, '270-80-900312',  'ISA(중개형)',       '신한투자증권', 536082, 'ISA',     FALSE);

-- 카드
INSERT INTO cards (customer_id, card_name, card_number, monthly_usage) VALUES
(1, '신한은행 Welpro', '9411-****-****-7823', 230340);

-- 거래내역 (최근 30일치 — 시나리오 1에서 7일 → 30일 변경 시연용)
INSERT INTO transactions (account_id, tx_type, amount, balance_after, description, counterparty, category, created_at) VALUES
-- 오늘
(1, 'WITHDRAW',  4500,    11234510, '스타벅스 강남점',        '스타벅스',     'FOOD',      NOW() - INTERVAL '0 days'),
-- 1일 전
(1, 'DEPOSIT',   3200000, 11239010, '4월 급여',              '신한은행',     'SALARY',    NOW() - INTERVAL '1 days'),
-- 2일 전
(1, 'WITHDRAW',  12300,   8039010,  '네이버페이 결제',        '네이버',       'SHOPPING',  NOW() - INTERVAL '2 days'),
-- 3일 전
(1, 'WITHDRAW',  1250,    8051310,  '서울 지하철',            '서울교통공사',  'TRANSPORT', NOW() - INTERVAL '3 days'),
(1, 'WITHDRAW',  8900,    8052560,  'GS25 편의점',           'GS리테일',     'FOOD',      NOW() - INTERVAL '3 days'),
-- 5일 전
(1, 'WITHDRAW',  45000,   8061460,  'CGV 영등포',            'CJ올리브',     'SHOPPING',  NOW() - INTERVAL '5 days'),
(1, 'TRANSFER',  500000,  8106460,  '김지원 → 신한급여통장',  '본인이체',     'TRANSFER',  NOW() - INTERVAL '5 days'),
-- 6일 전
(1, 'WITHDRAW',  15800,   8606460,  '배달의민족',             '우아한형제',   'FOOD',      NOW() - INTERVAL '6 days'),

-- ========= 7일 이후 (현재 쿼리에서 안 보이는 데이터) =========
-- 8일 전
(1, 'WITHDRAW',  89000,   8622260,  '쿠팡 주문',             '쿠팡',         'SHOPPING',  NOW() - INTERVAL '8 days'),
-- 10일 전
(1, 'WITHDRAW',  32000,   8711260,  '주유소 SK에너지',        'SK에너지',     'TRANSPORT', NOW() - INTERVAL '10 days'),
-- 12일 전
(1, 'DEPOSIT',   3200000, 8743260,  '3월 급여',              '신한은행',     'SALARY',    NOW() - INTERVAL '12 days'),
-- 15일 전
(1, 'WITHDRAW',  150000,  5543260,  '신한카드 자동이체',       '신한카드',     'TRANSFER',  NOW() - INTERVAL '15 days'),
-- 18일 전
(1, 'WITHDRAW',  67000,   5693260,  '이마트 장보기',          '이마트',       'FOOD',      NOW() - INTERVAL '18 days'),
-- 22일 전
(1, 'WITHDRAW',  25000,   5760260,  '택시비',                '카카오모빌',   'TRANSPORT', NOW() - INTERVAL '22 days'),
-- 28일 전
(1, 'WITHDRAW',  120000,  5785260,  '병원비 (건강검진)',       '서울아산',     'MEDICAL',   NOW() - INTERVAL '28 days');

-- 관심 종목
INSERT INTO stocks (customer_id, symbol, name, quantity, avg_price, current_price, change_pct, market, is_watchlist) VALUES
(1, 'TSLA',  '테슬라',         5,  350.00,  392.50,  -2.03, 'US', TRUE),
(1, 'NVDA',  '엔비디아',       3,  180.00,  202.06,  0.19,  'US', TRUE),
(1, 'PLTR',  '팔란티어 테크',   10, 120.00,  145.89,  -0.34, 'US', TRUE),
(1, '005930', '삼성전자',       50, 72000,   68500,   -1.42, 'KR', TRUE),
(1, '000270', '기아',           20, 95000,   102500,  1.15,  'KR', FALSE);
