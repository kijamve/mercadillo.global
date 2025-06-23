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
  `category_id` VARCHAR(36) NOT NULL,
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
  KEY `idx_products_category` (`category_id`),
  KEY `idx_products_status` (`status`),
  KEY `idx_products_price` (`price`),
  KEY `idx_products_rating` (`rating`),
  KEY `idx_products_stock` (`stock`),
  KEY `idx_products_kyc` (`kyc`),
  KEY `idx_products_from_company` (`from_company`),
  CONSTRAINT `fk_products_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
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

-- Insert sample user for testing
INSERT INTO `users` (`id`, `email`, `password`, `kyc_status`, `plan_slug`, `status`, `email_verified_at`, `created_at`, `updated_at`) 
VALUES 
('550e8400-e29b-41d4-a716-446655440001', 'user@example.com', '$2a$10$examplehashedpassword', 'approved', 'free', 'active', NULL, NOW(), NOW());

-- Add indexes for better performance
CREATE INDEX `idx_products_created_at` ON `products` (`created_at`);
CREATE INDEX `idx_products_updated_at` ON `products` (`updated_at`);
CREATE INDEX `idx_questions_created_at` ON `questions` (`created_at`);
CREATE INDEX `idx_reviews_created_at` ON `reviews` (`created_at`);

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

-- Additional indexes for advanced filtering
CREATE INDEX `idx_products_status` ON `products` (`status`);
CREATE INDEX `idx_products_price` ON `products` (`price`);
CREATE INDEX `idx_products_category` ON `products` (`category_id`);
CREATE INDEX `idx_products_rating` ON `products` (`rating`);
CREATE INDEX `idx_products_sold` ON `products` (`sold`);
