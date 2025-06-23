-- MySQL Schema for Mercadillo Global
-- Character Set: UTF8
-- Engine: InnoDB
-- Primary Keys: UUID (CHAR(36))

SET NAMES utf8;
SET character_set_client = utf8;

-- Users table
CREATE TABLE `users` (
  `id` CHAR(36) NOT NULL,
  `email` VARCHAR(255) NOT NULL,
  `password` VARCHAR(255) NOT NULL,
  `kyc_status` ENUM('pending','approved','rejected') NOT NULL DEFAULT 'pending',
  `plan_slug` VARCHAR(100) NOT NULL DEFAULT 'free',
  `status` ENUM('active','inactive','suspended') NOT NULL DEFAULT 'active',
  `company` BOOLEAN NOT NULL DEFAULT FALSE,
  `browser_settings` JSON,
  `user_details` JSON,
  `email_verified_at` TIMESTAMP NULL DEFAULT NULL,
  `deleted_at` TIMESTAMP NULL DEFAULT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_email` (`email`),
  KEY `idx_users_deleted_at` (`deleted_at`),
  KEY `idx_users_status` (`status`),
  KEY `idx_users_kyc_status` (`kyc_status`),
  KEY `idx_users_company` (`company`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- Products table
CREATE TABLE `products` (
  `id` CHAR(36) NOT NULL,
  `short_key` VARCHAR(20) NOT NULL,
  `slug` VARCHAR(255) NOT NULL,
  `user_id` CHAR(36) NOT NULL,
  `title` VARCHAR(100) NOT NULL,
  `price` DECIMAL(10,2) NOT NULL,
  `original_price` DECIMAL(10,2) NOT NULL DEFAULT 0,
  `currency_id` VARCHAR(3) NOT NULL DEFAULT 'USD',
  `price_type` ENUM('fixed','negotiable','per_hour','per_day','per_week','per_month','per_project') NOT NULL DEFAULT 'fixed',
  `images` JSON,
  `rating` DECIMAL(3,2) NOT NULL DEFAULT 0.00,
  `review_count` INT NOT NULL DEFAULT 0,
  `sold` INT NOT NULL DEFAULT 0,
  `stock` INT NOT NULL DEFAULT 0,
  `is_service` BOOLEAN NOT NULL DEFAULT FALSE,
  `free_shipping` BOOLEAN NOT NULL DEFAULT FALSE,
  `description` TEXT,
  `specifications` JSON,
  `search_content` TEXT COMMENT 'AI-generated optimized search content: title + category + key specs + keywords',
  `search_keywords` VARCHAR(500) COMMENT 'AI-generated comma-separated keywords for enhanced search',
  `status` ENUM('active','wait_for_ia','wait_for_human_review','pause','draft') NOT NULL DEFAULT 'draft',
  `kyc` BOOLEAN NOT NULL DEFAULT FALSE,
  `from_company` BOOLEAN NOT NULL DEFAULT FALSE,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_products_short_key` (`short_key`),
  UNIQUE KEY `idx_products_slug` (`slug`),
  KEY `fk_products_user` (`user_id`),
  KEY `idx_products_status` (`status`),
  KEY `idx_products_price` (`price`),
  KEY `idx_products_rating` (`rating`),
  KEY `idx_products_stock` (`stock`),
  KEY `idx_products_kyc` (`kyc`),
  KEY `idx_products_from_company` (`from_company`),
  CONSTRAINT `fk_products_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- Product categories table (many-to-many relationship)
CREATE TABLE `product_categories` (
  `id` CHAR(36) NOT NULL,
  `product_id` CHAR(36) NOT NULL,
  `category_id` VARCHAR(36) NOT NULL,
  `is_primary` BOOLEAN NOT NULL DEFAULT FALSE COMMENT 'Indicates if this is the primary category for the product',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_product_category_unique` (`product_id`, `category_id`),
  KEY `fk_product_categories_product` (`product_id`),
  KEY `idx_product_categories_category` (`category_id`),
  KEY `idx_product_categories_primary` (`is_primary`),
  CONSTRAINT `fk_product_categories_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- Warehouses table (owned by users)
CREATE TABLE `warehouses` (
  `id` CHAR(36) NOT NULL,
  `user_id` CHAR(36) NOT NULL,
  `name` VARCHAR(255) NOT NULL,
  `country` VARCHAR(2) NOT NULL,
  `state` VARCHAR(100) NOT NULL,
  `city` VARCHAR(100) NOT NULL,
  `address` TEXT NOT NULL,
  `postal_code` VARCHAR(20),
  `phone` VARCHAR(50),
  `email` VARCHAR(255),
  `is_active` BOOLEAN NOT NULL DEFAULT TRUE,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `fk_warehouses_user` (`user_id`),
  KEY `idx_warehouses_country` (`country`),
  KEY `idx_warehouses_state` (`state`),
  KEY `idx_warehouses_active` (`is_active`),
  CONSTRAINT `fk_warehouses_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- Product warehouses table (product stock in warehouses)
CREATE TABLE `product_warehouses` (
  `id` CHAR(36) NOT NULL,
  `product_id` CHAR(36) NOT NULL,
  `warehouse_id` CHAR(36) NOT NULL,
  `quantity` INT NOT NULL DEFAULT 0,
  `weight` DECIMAL(8,4) NOT NULL DEFAULT 0 COMMENT 'Weight in kg',
  `dimensions` JSON COMMENT 'JSON with length, width, height in cm',
  `specifications` JSON,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_product_warehouse_unique` (`product_id`, `warehouse_id`),
  KEY `fk_product_warehouses_product` (`product_id`),
  KEY `fk_product_warehouses_warehouse` (`warehouse_id`),
  KEY `idx_product_warehouses_quantity` (`quantity`),
  KEY `idx_product_warehouses_weight` (`weight`),
  CONSTRAINT `fk_product_warehouses_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `fk_product_warehouses_warehouse` FOREIGN KEY (`warehouse_id`) REFERENCES `warehouses` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- Shipping costs table
CREATE TABLE `shipping_costs` (
  `id` CHAR(36) NOT NULL,
  `product_warehouse_id` CHAR(36) NOT NULL,
  `country` VARCHAR(2) NOT NULL,
  `locations` JSON COMMENT 'Array of states/cities where this cost applies',
  `cost` DECIMAL(8,4) NOT NULL,
  `currency_id` VARCHAR(3) NOT NULL DEFAULT 'USD',
  `price_type` ENUM('fixed','per_kg') NOT NULL DEFAULT 'fixed',
  `min_weight` INT DEFAULT NULL COMMENT 'Minimum weight for this cost',
  `max_weight` INT DEFAULT NULL COMMENT 'Maximum weight for this cost',
  `estimated_days_min` INT DEFAULT NULL,
  `estimated_days_max` INT DEFAULT NULL,
  `is_active` BOOLEAN NOT NULL DEFAULT TRUE,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `fk_shipping_costs_product_warehouse` (`product_warehouse_id`),
  KEY `idx_shipping_costs_country` (`country`),
  KEY `idx_shipping_costs_price_type` (`price_type`),
  KEY `idx_shipping_costs_active` (`is_active`),
  CONSTRAINT `fk_shipping_costs_product_warehouse` FOREIGN KEY (`product_warehouse_id`) REFERENCES `product_warehouses` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- Product attributes table
CREATE TABLE `product_attributes` (
  `id` CHAR(36) NOT NULL,
  `product_id` CHAR(36) NOT NULL,
  `product_warehouse_id` CHAR(36) DEFAULT NULL,
  `attribute_slug` VARCHAR(100) NOT NULL,
  `value` JSON NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `fk_product_attributes_product` (`product_id`),
  KEY `idx_product_attributes_slug` (`attribute_slug`),
  KEY `fk_product_attributes_product_warehouse` (`product_warehouse_id`),
  CONSTRAINT `fk_product_attributes_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `fk_product_attributes_product_warehouse` FOREIGN KEY (`product_warehouse_id`) REFERENCES `product_warehouses` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- Questions table
CREATE TABLE `questions` (
  `id` CHAR(36) NOT NULL,
  `product_id` CHAR(36) NOT NULL,
  `question` TEXT NOT NULL,
  `answer` TEXT,
  `answered_by_ia` BOOLEAN NOT NULL DEFAULT FALSE,
  `helpful` INT NOT NULL DEFAULT 0,
  `status` ENUM('wait_for_ia','wait_for_human_review','hidden','answered') NOT NULL DEFAULT 'wait_for_ia',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `fk_questions_product` (`product_id`),
  KEY `idx_questions_status` (`status`),
  KEY `idx_questions_helpful` (`helpful`),
  CONSTRAINT `fk_questions_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- Reviews table
CREATE TABLE `reviews` (
  `id` CHAR(36) NOT NULL,
  `product_id` CHAR(36) NOT NULL,
  `name` VARCHAR(255) NOT NULL,
  `rating` TINYINT NOT NULL,
  `comment` TEXT,
  `helpful` INT NOT NULL DEFAULT 0,
  `status` ENUM('approved','wait_for_ia','wait_for_human_review','hidden') NOT NULL DEFAULT 'wait_for_ia',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `fk_reviews_product` (`product_id`),
  KEY `idx_reviews_status` (`status`),
  KEY `idx_reviews_rating` (`rating`),
  KEY `idx_reviews_helpful` (`helpful`),
  CONSTRAINT `fk_reviews_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `chk_reviews_rating` CHECK (`rating` >= 1 AND `rating` <= 5)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- Question votes table (pivot table for question helpful votes)
CREATE TABLE `question_votes` (
  `id` CHAR(36) NOT NULL,
  `user_id` CHAR(36) NOT NULL,
  `question_id` CHAR(36) NOT NULL,
  `vote` TINYINT NOT NULL COMMENT '1 for helpful, -1 for not helpful',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_question_votes_user_question` (`user_id`, `question_id`),
  KEY `fk_question_votes_user` (`user_id`),
  KEY `fk_question_votes_question` (`question_id`),
  KEY `idx_question_votes_vote` (`vote`),
  CONSTRAINT `fk_question_votes_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `fk_question_votes_question` FOREIGN KEY (`question_id`) REFERENCES `questions` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `chk_question_votes_vote` CHECK (`vote` IN (1, -1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- Review votes table (pivot table for review helpful votes)
CREATE TABLE `review_votes` (
  `id` CHAR(36) NOT NULL,
  `user_id` CHAR(36) NOT NULL,
  `review_id` CHAR(36) NOT NULL,
  `vote` TINYINT NOT NULL COMMENT '1 for helpful, -1 for not helpful',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_review_votes_user_review` (`user_id`, `review_id`),
  KEY `fk_review_votes_user` (`user_id`),
  KEY `fk_review_votes_review` (`review_id`),
  KEY `idx_review_votes_vote` (`vote`),
  CONSTRAINT `fk_review_votes_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `fk_review_votes_review` FOREIGN KEY (`review_id`) REFERENCES `reviews` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `chk_review_votes_vote` CHECK (`vote` IN (1, -1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- Add full-text search indexes for product search
-- Primary optimized search index using AI-generated content
CREATE FULLTEXT INDEX `idx_products_search_optimized` ON `products` (`search_content`, `search_keywords`);

-- Fallback index for basic search (title, description)
CREATE FULLTEXT INDEX `idx_products_search_basic` ON `products` (`title`, `description`);


-- Insert sample user for testing
INSERT INTO `users` (`id`, `email`, `password`, `kyc_status`, `plan_slug`, `status`, `email_verified_at`, `created_at`, `updated_at`) 
VALUES 
('550e8400-e29b-41d4-a716-K0001', 'user@example.com', '$2a$10$examplehashedpassword', 'approved', 'free', 'active', NULL, NOW(), NOW());


-- Insert additional dummy users
INSERT INTO `users` (`id`, `email`, `password`, `kyc_status`, `plan_slug`, `status`, `email_verified_at`, `created_at`, `updated_at`) VALUES
('550e8400-e29b-41d4-a716-K0002', 'maria@example.com', '$2a$10$examplehashedpassword2', 'approved', 'premium', 'active', NOW(), NOW(), NOW()),
('550e8400-e29b-41d4-a716-K0003', 'carlos@example.com', '$2a$10$examplehashedpassword3', 'approved', 'free', 'active', NOW(), NOW(), NOW()),
('550e8400-e29b-41d4-a716-K0004', 'ana@example.com', '$2a$10$examplehashedpassword4', 'pending', 'free', 'active', NULL, NOW(), NOW()),
('550e8400-e29b-41d4-a716-K0005', 'tech@company.com', '$2a$10$examplehashedpassword5', 'approved', 'business', 'active', NOW(), NOW(), NOW()),
('550e8400-e29b-41d4-a716-K0006', 'review@user.com', '$2a$10$examplehashedpassword6', 'approved', 'free', 'active', NOW(), NOW(), NOW());

-- Insert dummy products
INSERT INTO `products` (`id`, `short_key`, `slug`, `user_id`, `title`, `price`, `original_price`, `currency_id`, `images`, `rating`, `review_count`, `sold`, `stock`, `is_service`, `free_shipping`, `description`, `specifications`, `search_content`, `search_keywords`, `status`, `kyc`, `from_company`, `created_at`, `updated_at`) VALUES
('prod-8400-e29b-41d4-a716-K0001', 'LAP001', 'laptop-gaming-asus-rog', '550e8400-e29b-41d4-a716-K0005', 'Laptop Gaming ASUS ROG Strix G15', 129999, 149999, 'USD', '["laptop1.jpg", "laptop2.jpg", "laptop3.jpg"]', 4.50, 12, 25, 15, false, true, 'Laptop gaming de alto rendimiento con procesador AMD Ryzen 7, tarjeta gráfica NVIDIA RTX 3060, 16GB RAM y 512GB SSD. Perfecta para gaming y trabajo profesional.', '[{"name":"Procesador","value":"AMD Ryzen 7 5800H"},{"name":"RAM","value":"16GB DDR4"},{"name":"Almacenamiento","value":"512GB SSD"},{"name":"Tarjeta Gráfica","value":"NVIDIA RTX 3060"}]', 'Laptop Gaming ASUS ROG Strix G15 Tecnología AMD Ryzen RTX gaming computadora portatil alto rendimiento', 'laptop, gaming, ASUS, ROG, AMD, Ryzen, NVIDIA, RTX', 'active', false, true, NOW(), NOW()),

('prod-8400-e29b-41d4-a716-K0002', 'MOV001', 'smartphone-samsung-galaxy-s23', '550e8400-e29b-41d4-a716-K0002', 'Samsung Galaxy S23 128GB', 79999, 89999, 'USD', '["phone1.jpg", "phone2.jpg"]', 4.20, 8, 42, 28, false, true, 'Smartphone Samsung Galaxy S23 con cámara de 50MP, pantalla Dynamic AMOLED 2X de 6.1 pulgadas, procesador Snapdragon 8 Gen 2 y batería de larga duración.', '[{"name":"Pantalla","value":"6.1 pulgadas Dynamic AMOLED 2X"},{"name":"Cámara","value":"50MP + 12MP + 10MP"},{"name":"Procesador","value":"Snapdragon 8 Gen 2"},{"name":"Batería","value":"3900 mAh"}]', 'Samsung Galaxy S23 128GB Telecomunicaciones smartphone celular cámara Android', 'Samsung, Galaxy, S23, smartphone, Android, cámara', 'active', false, false, NOW(), NOW()),

('prod-8400-e29b-41d4-a716-K0003', 'SER001', 'servicio-desarrollo-web', '550e8400-e29b-41d4-a716-K0003', 'Desarrollo de Página Web Profesional', 50000, 0, 'USD', '["web1.jpg", "web2.jpg", "web3.jpg"]', 4.80, 15, 35, 0, true, false, 'Servicio profesional de desarrollo web con diseño responsive, optimización SEO, integración de base de datos y panel administrativo. Incluye hosting por 1 año.', '[{"name":"Tecnologías","value":"HTML5, CSS3, JavaScript, PHP"},{"name":"Tiempo de entrega","value":"2-4 semanas"},{"name":"Garantía","value":"6 meses"},{"name":"Revisiones","value":"3 revisiones incluidas"}]', 'Desarrollo Página Web Profesional Servicios programación diseño responsive SEO', 'desarrollo web, diseño, responsive, SEO, programación', 'active', false, false, NOW(), NOW()),

('prod-8400-e29b-41d4-a716-K0004', 'DEP001', 'bicicleta-mountain-bike-trek', '550e8400-e29b-41d4-a716-K0001', 'Bicicleta Mountain Bike Trek Marlin 7', 85000, 95000, 'USD', '["bike1.jpg", "bike2.jpg", "bike3.jpg", "bike4.jpg"]', 4.30, 6, 18, 12, false, false, 'Bicicleta de montaña Trek Marlin 7 con cuadro de aluminio, suspensión delantera, cambios Shimano de 21 velocidades y frenos de disco mecánicos.', '[{"name":"Cuadro","value":"Aluminio Alpha Silver"},{"name":"Suspensión","value":"Delantera SR Suntour XCE"},{"name":"Cambios","value":"Shimano 21 velocidades"},{"name":"Frenos","value":"Disco mecánico"}]', 'Bicicleta Mountain Bike Trek Marlin 7 Deportes ciclismo montaña aluminio Shimano', 'bicicleta, mountain bike, Trek, ciclismo, deportes', 'active', false, false, NOW(), NOW()),

('prod-8400-e29b-41d4-a716-K0005', 'MAS001', 'alimento-perros-royal-canin', '550e8400-e29b-41d4-a716-K0004', 'Royal Canin Adult Perros Adultos 15kg', 12500, 14000, 'USD', '["dog_food1.jpg", "dog_food2.jpg"]', 4.60, 22, 156, 45, false, true, 'Alimento balanceado premium para perros adultos de todas las razas. Fórmula completa y equilibrada que contribuye a la salud digestiva y fortalece el sistema inmunitario.', '[{"name":"Peso","value":"15kg"},{"name":"Edad","value":"Perros adultos (1-7 años)"},{"name":"Proteína","value":"22% mínimo"},{"name":"Grasa","value":"12% mínimo"}]', 'Royal Canin Adult Perros Adultos 15kg Mascotas alimento balanceado premium digestivo', 'Royal Canin, alimento perros, premium, adultos, balanceado', 'active', false, false, NOW(), NOW()),

('prod-8400-e29b-41d4-a716-K0006', 'HOG001', 'aspiradora-dyson-v15-detect', '550e8400-e29b-41d4-a716-K0005', 'Aspiradora Dyson V15 Detect Inalámbrica', 65000, 75000, 'USD', '["vacuum1.jpg", "vacuum2.jpg", "vacuum3.jpg"]', 4.70, 9, 31, 8, false, true, 'Aspiradora inalámbrica Dyson V15 Detect con tecnología láser que revela polvo microscópico, motor Hyperdymium y hasta 60 minutos de autonomía.', '[{"name":"Autonomía","value":"Hasta 60 minutos"},{"name":"Tecnología","value":"Láser Detect"},{"name":"Motor","value":"Hyperdymium"},{"name":"Filtración","value":"HEPA completa"}]', 'Aspiradora Dyson V15 Detect Inalámbrica Aparatos Limpieza láser tecnología HEPA', 'Dyson, aspiradora, inalámbrica, láser, limpieza', 'active', false, true, NOW(), NOW());

-- Insert product categories relationships
INSERT INTO `product_categories` (`id`, `product_id`, `category_id`, `is_primary`, `created_at`, `updated_at`) VALUES
('cat-8400-e29b-41d4-a716-K0001', 'prod-8400-e29b-41d4-a716-K0001', 'computadores-portatiles', true, NOW(), NOW()),
('cat-8400-e29b-41d4-a716-K0002', 'prod-8400-e29b-41d4-a716-K0001', 'accesorios-gaming', false, NOW(), NOW()),
('cat-8400-e29b-41d4-a716-K0003', 'prod-8400-e29b-41d4-a716-K0002', 'celulares', true, NOW(), NOW()),
('cat-8400-e29b-41d4-a716-K0004', 'prod-8400-e29b-41d4-a716-K0003', 'desarrollo-y-programacion', true, NOW(), NOW()),
('cat-8400-e29b-41d4-a716-K0005', 'prod-8400-e29b-41d4-a716-K0004', 'ciclismo', true, NOW(), NOW()),
('cat-8400-e29b-41d4-a716-K0006', 'prod-8400-e29b-41d4-a716-K0005', 'alimentos-mascotas', true, NOW(), NOW()),
('cat-8400-e29b-41d4-a716-K0007', 'prod-8400-e29b-41d4-a716-K0006', 'limpieza', true, NOW(), NOW());

-- Insert dummy reviews
INSERT INTO `reviews` (`id`, `product_id`, `name`, `rating`, `comment`, `helpful`, `status`, `created_at`, `updated_at`) VALUES
-- Reviews for Laptop Gaming ASUS
('rev-8400-e29b-41d4-a716-K0001', 'prod-8400-e29b-41d4-a716-K0001', 'Miguel Torres', 5, 'Excelente laptop gaming! El rendimiento es increíble, corre todos los juegos en configuración alta sin problemas. La pantalla es hermosa y el teclado RGB es genial.', 8, 'approved', NOW() - INTERVAL 15 DAY, NOW() - INTERVAL 15 DAY),
('rev-8400-e29b-41d4-a716-K0002', 'prod-8400-e29b-41d4-a716-K0001', 'Carolina Ruiz', 4, 'Muy buena laptop, aunque se calienta un poco durante sesiones largas de gaming. Por lo demás, está perfecta para trabajo y entretenimiento.', 5, 'approved', NOW() - INTERVAL 10 DAY, NOW() - INTERVAL 10 DAY),
('rev-8400-e29b-41d4-a716-K0003', 'prod-8400-e29b-41d4-a716-K0001', 'Andrés López', 5, 'Llegó súper rápido y en perfectas condiciones. El vendedor muy profesional. La laptop es exactamente lo que esperaba.', 12, 'approved', NOW() - INTERVAL 7 DAY, NOW() - INTERVAL 7 DAY),

-- Reviews for Samsung Galaxy S23
('rev-8400-e29b-41d4-a716-K0004', 'prod-8400-e29b-41d4-a716-K0002', 'Lucía Mendoza', 4, 'Buen teléfono, las cámaras toman fotos espectaculares y la batería dura todo el día. La pantalla es muy nítida.', 6, 'approved', NOW() - INTERVAL 12 DAY, NOW() - INTERVAL 12 DAY),
('rev-8400-e29b-41d4-a716-K0005', 'prod-8400-e29b-41d4-a716-K0002', 'Roberto Vega', 4, 'Excelente relación calidad-precio. El teléfono es rápido y las fotos nocturnas son impresionantes.', 4, 'approved', NOW() - INTERVAL 8 DAY, NOW() - INTERVAL 8 DAY),

-- Reviews for Desarrollo Web Service
('rev-8400-e29b-41d4-a716-K0006', 'prod-8400-e29b-41d4-a716-K0003', 'Elena Morales', 5, 'Carlos hizo un trabajo excepcional con mi sitio web. Muy profesional, cumplió con los tiempos y el resultado superó mis expectativas.', 15, 'approved', NOW() - INTERVAL 20 DAY, NOW() - INTERVAL 20 DAY),
('rev-8400-e29b-41d4-a716-K0007', 'prod-8400-e29b-41d4-a716-K0003', 'Francisco Díaz', 5, 'Servicio de primera calidad. Mi tienda online funciona perfectamente y las ventas han aumentado considerablemente.', 11, 'approved', NOW() - INTERVAL 14 DAY, NOW() - INTERVAL 14 DAY),
('rev-8400-e29b-41d4-a716-K0008', 'prod-8400-e29b-41d4-a716-K0003', 'Isabel Castro', 4, 'Buen trabajo, aunque hubo algunos retrasos menores. El resultado final es muy bueno y el soporte post-entrega excelente.', 7, 'approved', NOW() - INTERVAL 9 DAY, NOW() - INTERVAL 9 DAY),

-- Reviews for Mountain Bike Trek
('rev-8400-e29b-41d4-a716-K0009', 'prod-8400-e29b-41d4-a716-K0004', 'Diego Ramírez', 4, 'Bicicleta sólida y bien construida. Los componentes Shimano funcionan suavemente. Perfecta para trails de montaña.', 3, 'approved', NOW() - INTERVAL 11 DAY, NOW() - INTERVAL 11 DAY),
('rev-8400-e29b-41d4-a716-K0010', 'prod-8400-e29b-41d4-a716-K0004', 'Valentina Cruz', 5, 'Me encanta esta bicicleta! Es ligera, resistente y muy cómoda para rutas largas. Excelente compra.', 8, 'approved', NOW() - INTERVAL 6 DAY, NOW() - INTERVAL 6 DAY),

-- Reviews for Royal Canin Dog Food
('rev-8400-e29b-41d4-a716-K0011', 'prod-8400-e29b-41d4-a716-K0005', 'Patricia Herrera', 5, 'Mi perro adora este alimento. Desde que lo cambié a Royal Canin, su pelaje está más brillante y tiene más energía.', 14, 'approved', NOW() - INTERVAL 18 DAY, NOW() - INTERVAL 18 DAY),
('rev-8400-e29b-41d4-a716-K0012', 'prod-8400-e29b-41d4-a716-K0005', 'Raúl Jiménez', 4, 'Buen alimento, aunque un poco costoso. Mi golden retriever lo digiere muy bien y no ha tenido problemas estomacales.', 6, 'approved', NOW() - INTERVAL 13 DAY, NOW() - INTERVAL 13 DAY),

-- Reviews for Dyson Vacuum
('rev-8400-e29b-41d4-a716-K0013', 'prod-8400-e29b-41d4-a716-K0006', 'Camila Vargas', 5, 'Increíble aspiradora! La tecnología láser realmente revela polvo que no sabía que existía. Vale cada centavo.', 10, 'approved', NOW() - INTERVAL 16 DAY, NOW() - INTERVAL 16 DAY),
('rev-8400-e29b-41d4-a716-K0014', 'prod-8400-e29b-41d4-a716-K0006', 'Jorge Medina', 4, 'Muy buena aspiradora, potente y práctica. La batería dura lo prometido. Un poco ruidosa pero nada grave.', 4, 'approved', NOW() - INTERVAL 5 DAY, NOW() - INTERVAL 5 DAY);

-- Insert review votes
INSERT INTO `review_votes` (`id`, `user_id`, `review_id`, `vote`, `created_at`, `updated_at`) VALUES
-- Votes for laptop reviews
('vote-8400-e29b-41d4-a716-K0001', '550e8400-e29b-41d4-a716-K0002', 'rev-8400-e29b-41d4-a716-K0001', 1, NOW() - INTERVAL 14 DAY, NOW() - INTERVAL 14 DAY),
('vote-8400-e29b-41d4-a716-K0002', '550e8400-e29b-41d4-a716-K0003', 'rev-8400-e29b-41d4-a716-K0001', 1, NOW() - INTERVAL 13 DAY, NOW() - INTERVAL 13 DAY),
('vote-8400-e29b-41d4-a716-K0003', '550e8400-e29b-41d4-a716-K0004', 'rev-8400-e29b-41d4-a716-K0001', 1, NOW() - INTERVAL 12 DAY, NOW() - INTERVAL 12 DAY),
('vote-8400-e29b-41d4-a716-K0004', '550e8400-e29b-41d4-a716-K0006', 'rev-8400-e29b-41d4-a716-K0002', 1, NOW() - INTERVAL 9 DAY, NOW() - INTERVAL 9 DAY),
('vote-8400-e29b-41d4-a716-K0005', '550e8400-e29b-41d4-a716-K0001', 'rev-8400-e29b-41d4-a716-K0003', 1, NOW() - INTERVAL 6 DAY, NOW() - INTERVAL 6 DAY),

-- Votes for phone reviews  
('vote-8400-e29b-41d4-a716-K0006', '550e8400-e29b-41d4-a716-K0003', 'rev-8400-e29b-41d4-a716-K0004', 1, NOW() - INTERVAL 11 DAY, NOW() - INTERVAL 11 DAY),
('vote-8400-e29b-41d4-a716-K0007', '550e8400-e29b-41d4-a716-K0001', 'rev-8400-e29b-41d4-a716-K0005', 1, NOW() - INTERVAL 7 DAY, NOW() - INTERVAL 7 DAY),

-- Votes for web service reviews
('vote-8400-e29b-41d4-a716-K0008', '550e8400-e29b-41d4-a716-K0001', 'rev-8400-e29b-41d4-a716-K0006', 1, NOW() - INTERVAL 19 DAY, NOW() - INTERVAL 19 DAY),
('vote-8400-e29b-41d4-a716-K0009', '550e8400-e29b-41d4-a716-K0002', 'rev-8400-e29b-41d4-a716-K0007', 1, NOW() - INTERVAL 13 DAY, NOW() - INTERVAL 13 DAY),
('vote-8400-e29b-41d4-a716-K0010', '550e8400-e29b-41d4-a716-K0004', 'rev-8400-e29b-41d4-a716-K0008', 1, NOW() - INTERVAL 8 DAY, NOW() - INTERVAL 8 DAY),

-- Votes for bike reviews
('vote-8400-e29b-41d4-a716-K0011', '550e8400-e29b-41d4-a716-K0002', 'rev-8400-e29b-41d4-a716-K0010', 1, NOW() - INTERVAL 5 DAY, NOW() - INTERVAL 5 DAY),

-- Votes for dog food reviews
('vote-8400-e29b-41d4-a716-K0012', '550e8400-e29b-41d4-a716-K0001', 'rev-8400-e29b-41d4-a716-K0011', 1, NOW() - INTERVAL 17 DAY, NOW() - INTERVAL 17 DAY),
('vote-8400-e29b-41d4-a716-K0013', '550e8400-e29b-41d4-a716-K0003', 'rev-8400-e29b-41d4-a716-K0012', 1, NOW() - INTERVAL 12 DAY, NOW() - INTERVAL 12 DAY),

-- Votes for vacuum reviews
('vote-8400-e29b-41d4-a716-K0014', '550e8400-e29b-41d4-a716-K0002', 'rev-8400-e29b-41d4-a716-K0013', 1, NOW() - INTERVAL 15 DAY, NOW() - INTERVAL 15 DAY),
('vote-8400-e29b-41d4-a716-K0015', '550e8400-e29b-41d4-a716-K0003', 'rev-8400-e29b-41d4-a716-K0014', 1, NOW() - INTERVAL 4 DAY, NOW() - INTERVAL 4 DAY);

-- Insert dummy questions
INSERT INTO `questions` (`id`, `product_id`, `question`, `answer`, `answered_by_ia`, `helpful`, `status`, `created_at`, `updated_at`) VALUES
-- Questions for Laptop Gaming ASUS
('ques-8400-e29b-41d4-a716-K0001', 'prod-8400-e29b-41d4-a716-K0001', '¿Viene con Windows preinstalado?', 'Sí, la laptop viene con Windows 11 Home preinstalado y activado. También incluye el software ROG Armoury Crate para personalizar iluminación y rendimiento.', false, 5, 'answered', NOW() - INTERVAL 8 DAY, NOW() - INTERVAL 8 DAY),
('ques-8400-e29b-41d4-a716-K0002', 'prod-8400-e29b-41d4-a716-K0001', '¿Qué juegos puede correr sin problemas?', 'Esta laptop puede ejecutar juegos AAA actuales como Cyberpunk 2077, Call of Duty, Apex Legends en configuración alta-ultra con fps estables. Es perfecta para gaming competitivo y casual.', false, 8, 'answered', NOW() - INTERVAL 5 DAY, NOW() - INTERVAL 5 DAY),

-- Questions for Samsung Galaxy S23
('ques-8400-e29b-41d4-a716-K0003', 'prod-8400-e29b-41d4-a716-K0002', '¿Incluye cargador en la caja?', 'El Samsung Galaxy S23 incluye cable USB-C pero no incluye adaptador de pared. El teléfono es compatible con carga rápida de 25W y carga inalámbrica.', false, 6, 'answered', NOW() - INTERVAL 6 DAY, NOW() - INTERVAL 6 DAY),
('ques-8400-e29b-41d4-a716-K0004', 'prod-8400-e29b-41d4-a716-K0002', '¿La cámara nocturna es realmente buena?', 'Sí, las fotos nocturnas son excelentes gracias al modo Night Mode mejorado y el sensor principal de 50MP con estabilización óptica. Los resultados son muy nítidos incluso en condiciones de poca luz.', false, 9, 'answered', NOW() - INTERVAL 3 DAY, NOW() - INTERVAL 3 DAY),

-- Questions for Web Development Service
('ques-8400-e29b-41d4-a716-K0005', 'prod-8400-e29b-41d4-a716-K0003', '¿Qué incluye exactamente el servicio?', 'El servicio incluye: diseño personalizado responsive, hasta 10 páginas, formulario de contacto, optimización SEO básica, integración con redes sociales, hosting por 1 año y 3 rondas de revisiones.', false, 12, 'answered', NOW() - INTERVAL 4 DAY, NOW() - INTERVAL 4 DAY),

-- Questions for Mountain Bike Trek
('ques-8400-e29b-41d4-a716-K0006', 'prod-8400-e29b-41d4-a716-K0004', '¿Qué talla de bicicleta necesito?', 'Para elegir la talla correcta, considera tu altura: S (155-168cm), M (168-178cm), L (178-188cm), XL (188-198cm). También ofrecemos asesoría personalizada para encontrar la talla perfecta.', false, 4, 'answered', NOW() - INTERVAL 7 DAY, NOW() - INTERVAL 7 DAY),

-- Questions for Royal Canin Dog Food
('ques-8400-e29b-41d4-a716-K0007', 'prod-8400-e29b-41d4-a716-K0005', '¿Es adecuado para todas las razas?', 'Sí, Royal Canin Adult está formulado para perros adultos de todas las razas y tamaños. Contiene los nutrientes esenciales para mantener la salud general de tu mascota.', false, 7, 'answered', NOW() - INTERVAL 9 DAY, NOW() - INTERVAL 9 DAY),

-- Questions for Dyson Vacuum
('ques-8400-e29b-41d4-a716-K0008', 'prod-8400-e29b-41d4-a716-K0006', '¿Funciona bien en alfombras gruesas?', 'Absolutamente. La Dyson V15 Detect tiene un cabezal motorizado específico para alfombras que ajusta automáticamente la potencia de succión. Detecta el tipo de superficie y optimiza el rendimiento.', false, 6, 'answered', NOW() - INTERVAL 2 DAY, NOW() - INTERVAL 2 DAY);

-- Insert question votes
INSERT INTO `question_votes` (`id`, `user_id`, `question_id`, `vote`, `created_at`, `updated_at`) VALUES
-- Votes for laptop questions
('qvote-8400-e29b-41d4-a716-K0001', '550e8400-e29b-41d4-a716-K0002', 'ques-8400-e29b-41d4-a716-K0001', 1, NOW() - INTERVAL 7 DAY, NOW() - INTERVAL 7 DAY),
('qvote-8400-e29b-41d4-a716-K0002', '550e8400-e29b-41d4-a716-K0003', 'ques-8400-e29b-41d4-a716-K0002', 1, NOW() - INTERVAL 4 DAY, NOW() - INTERVAL 4 DAY),
('qvote-8400-e29b-41d4-a716-K0003', '550e8400-e29b-41d4-a716-K0004', 'ques-8400-e29b-41d4-a716-K0002', 1, NOW() - INTERVAL 3 DAY, NOW() - INTERVAL 3 DAY),

-- Votes for phone questions
('qvote-8400-e29b-41d4-a716-K0004', '550e8400-e29b-41d4-a716-K0001', 'ques-8400-e29b-41d4-a716-K0003', 1, NOW() - INTERVAL 5 DAY, NOW() - INTERVAL 5 DAY),
('qvote-8400-e29b-41d4-a716-K0005', '550e8400-e29b-41d4-a716-K0005', 'ques-8400-e29b-41d4-a716-K0004', 1, NOW() - INTERVAL 2 DAY, NOW() - INTERVAL 2 DAY),

-- Votes for web service questions
('qvote-8400-e29b-41d4-a716-K0006', '550e8400-e29b-41d4-a716-K0001', 'ques-8400-e29b-41d4-a716-K0005', 1, NOW() - INTERVAL 3 DAY, NOW() - INTERVAL 3 DAY),
('qvote-8400-e29b-41d4-a716-K0007', '550e8400-e29b-41d4-a716-K0002', 'ques-8400-e29b-41d4-a716-K0005', 1, NOW() - INTERVAL 2 DAY, NOW() - INTERVAL 2 DAY),

-- Votes for bike questions
('qvote-8400-e29b-41d4-a716-K0008', '550e8400-e29b-41d4-a716-K0002', 'ques-8400-e29b-41d4-a716-K0006', 1, NOW() - INTERVAL 6 DAY, NOW() - INTERVAL 6 DAY),

-- Votes for dog food questions
('qvote-8400-e29b-41d4-a716-K0009', '550e8400-e29b-41d4-a716-K0003', 'ques-8400-e29b-41d4-a716-K0007', 1, NOW() - INTERVAL 8 DAY, NOW() - INTERVAL 8 DAY),

-- Votes for vacuum questions
('qvote-8400-e29b-41d4-a716-K0010', '550e8400-e29b-41d4-a716-K0004', 'ques-8400-e29b-41d4-a716-K0008', 1, NOW() - INTERVAL 1 DAY, NOW() - INTERVAL 1 DAY);
