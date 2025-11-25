-- ============================================
-- Go Blockchain Wallet 数据库建表语句
-- ============================================
-- 支持 PostgreSQL 和 MySQL
-- ============================================

-- ============================================
-- 1. 资产配置表 (assets)
-- ============================================
-- PostgreSQL 版本
CREATE TABLE IF NOT EXISTS assets (
    symbol VARCHAR(50) PRIMARY KEY,
    chain VARCHAR(20) NOT NULL,
    decimals SMALLINT NOT NULL DEFAULT 0,
    token_addr VARCHAR(255) DEFAULT '',
    tags JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_assets_chain ON assets(chain);

-- MySQL 版本
/*
CREATE TABLE IF NOT EXISTS assets (
    symbol VARCHAR(50) PRIMARY KEY,
    chain VARCHAR(20) NOT NULL,
    decimals TINYINT UNSIGNED NOT NULL DEFAULT 0,
    token_addr VARCHAR(255) DEFAULT '',
    tags JSON DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_assets_chain ON assets(chain);
*/

-- ============================================
-- 2. 钱包账户表 (wallet_accounts)
-- ============================================
-- PostgreSQL 版本
CREATE TABLE IF NOT EXISTS wallet_accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    asset_symbol VARCHAR(50) NOT NULL,
    address VARCHAR(255) NOT NULL,
    chain VARCHAR(20) NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, asset_symbol)
);

CREATE INDEX idx_wallet_accounts_user_id ON wallet_accounts(user_id);
CREATE INDEX idx_wallet_accounts_asset_symbol ON wallet_accounts(asset_symbol);
CREATE INDEX idx_wallet_accounts_address ON wallet_accounts(address);
CREATE INDEX idx_wallet_accounts_address_asset ON wallet_accounts(address, asset_symbol);

-- MySQL 版本
/*
CREATE TABLE IF NOT EXISTS wallet_accounts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    asset_symbol VARCHAR(50) NOT NULL,
    address VARCHAR(255) NOT NULL,
    chain VARCHAR(20) NOT NULL,
    metadata JSON DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_asset (user_id, asset_symbol)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_wallet_accounts_user_id ON wallet_accounts(user_id);
CREATE INDEX idx_wallet_accounts_asset_symbol ON wallet_accounts(asset_symbol);
CREATE INDEX idx_wallet_accounts_address ON wallet_accounts(address);
CREATE INDEX idx_wallet_accounts_address_asset ON wallet_accounts(address, asset_symbol);
*/

-- ============================================
-- 3. 余额表 (balances)
-- ============================================
-- PostgreSQL 版本
CREATE TABLE IF NOT EXISTS balances (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    asset_symbol VARCHAR(50) NOT NULL,
    available NUMERIC(78, 0) NOT NULL DEFAULT 0,  -- big.Int 使用 NUMERIC(78,0) 存储
    frozen NUMERIC(78, 0) NOT NULL DEFAULT 0,
    pending NUMERIC(78, 0) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, asset_symbol)
);

CREATE INDEX idx_balances_user_id ON balances(user_id);
CREATE INDEX idx_balances_asset_symbol ON balances(asset_symbol);

-- MySQL 版本
/*
CREATE TABLE IF NOT EXISTS balances (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    asset_symbol VARCHAR(50) NOT NULL,
    available DECIMAL(78, 0) NOT NULL DEFAULT 0,  -- big.Int 使用 DECIMAL(78,0) 存储
    frozen DECIMAL(78, 0) NOT NULL DEFAULT 0,
    pending DECIMAL(78, 0) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_asset (user_id, asset_symbol)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_balances_user_id ON balances(user_id);
CREATE INDEX idx_balances_asset_symbol ON balances(asset_symbol);
*/

-- ============================================
-- 4. 充值记录表 (deposits)
-- ============================================
-- PostgreSQL 版本
CREATE TABLE IF NOT EXISTS deposits (
    tx_hash VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    chain VARCHAR(20) NOT NULL,
    asset_symbol VARCHAR(50) NOT NULL,
    amount NUMERIC(78, 0) NOT NULL,
    from_address VARCHAR(255) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    block_height BIGINT NOT NULL,
    confirmations BIGINT NOT NULL DEFAULT 0,
    required_confirmations BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    observed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    credited_at TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_deposits_user_id ON deposits(user_id);
CREATE INDEX idx_deposits_chain ON deposits(chain);
CREATE INDEX idx_deposits_asset_symbol ON deposits(asset_symbol);
CREATE INDEX idx_deposits_status ON deposits(status);
CREATE INDEX idx_deposits_block_height ON deposits(chain, block_height);
CREATE INDEX idx_deposits_to_address ON deposits(to_address);
CREATE INDEX idx_deposits_observed_at ON deposits(observed_at);

-- MySQL 版本
/*
CREATE TABLE IF NOT EXISTS deposits (
    tx_hash VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    chain VARCHAR(20) NOT NULL,
    asset_symbol VARCHAR(50) NOT NULL,
    amount DECIMAL(78, 0) NOT NULL,
    from_address VARCHAR(255) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    block_height BIGINT UNSIGNED NOT NULL,
    confirmations BIGINT UNSIGNED NOT NULL DEFAULT 0,
    required_confirmations BIGINT UNSIGNED NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    observed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    credited_at TIMESTAMP NULL,
    metadata JSON DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_deposits_user_id ON deposits(user_id);
CREATE INDEX idx_deposits_chain ON deposits(chain);
CREATE INDEX idx_deposits_asset_symbol ON deposits(asset_symbol);
CREATE INDEX idx_deposits_status ON deposits(status);
CREATE INDEX idx_deposits_block_height ON deposits(chain, block_height);
CREATE INDEX idx_deposits_to_address ON deposits(to_address);
CREATE INDEX idx_deposits_observed_at ON deposits(observed_at);
*/

-- ============================================
-- 5. 提现记录表 (withdrawals)
-- ============================================
-- PostgreSQL 版本
CREATE TABLE IF NOT EXISTS withdrawals (
    id VARCHAR(100) PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    asset_symbol VARCHAR(50) NOT NULL,
    chain VARCHAR(20) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    amount NUMERIC(78, 0) NOT NULL,
    fee NUMERIC(78, 0) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'requested',
    risk_score DOUBLE PRECISION DEFAULT 0,
    risk_remarks TEXT,
    raw_tx BYTEA,
    tx_hash VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_withdrawals_user_id ON withdrawals(user_id);
CREATE INDEX idx_withdrawals_asset_symbol ON withdrawals(asset_symbol);
CREATE INDEX idx_withdrawals_chain ON withdrawals(chain);
CREATE INDEX idx_withdrawals_status ON withdrawals(status);
CREATE INDEX idx_withdrawals_tx_hash ON withdrawals(tx_hash);
CREATE INDEX idx_withdrawals_created_at ON withdrawals(created_at);

-- MySQL 版本
/*
CREATE TABLE IF NOT EXISTS withdrawals (
    id VARCHAR(100) PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    asset_symbol VARCHAR(50) NOT NULL,
    chain VARCHAR(20) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    amount DECIMAL(78, 0) NOT NULL,
    fee DECIMAL(78, 0) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'requested',
    risk_score DOUBLE DEFAULT 0,
    risk_remarks TEXT,
    raw_tx BLOB,
    tx_hash VARCHAR(255),
    metadata JSON DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_withdrawals_user_id ON withdrawals(user_id);
CREATE INDEX idx_withdrawals_asset_symbol ON withdrawals(asset_symbol);
CREATE INDEX idx_withdrawals_chain ON withdrawals(chain);
CREATE INDEX idx_withdrawals_status ON withdrawals(status);
CREATE INDEX idx_withdrawals_tx_hash ON withdrawals(tx_hash);
CREATE INDEX idx_withdrawals_created_at ON withdrawals(created_at);
*/

-- ============================================
-- 6. 区块信息表 (blocks)
-- ============================================
-- PostgreSQL 版本
CREATE TABLE IF NOT EXISTS blocks (
    id BIGSERIAL PRIMARY KEY,
    chain VARCHAR(20) NOT NULL,
    height BIGINT NOT NULL,
    hash VARCHAR(255) NOT NULL,
    parent_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(chain, height)
);

CREATE INDEX idx_blocks_chain ON blocks(chain);
CREATE INDEX idx_blocks_height ON blocks(chain, height);
CREATE INDEX idx_blocks_hash ON blocks(hash);

-- MySQL 版本
/*
CREATE TABLE IF NOT EXISTS blocks (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    chain VARCHAR(20) NOT NULL,
    height BIGINT UNSIGNED NOT NULL,
    hash VARCHAR(255) NOT NULL,
    parent_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_chain_height (chain, height)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_blocks_chain ON blocks(chain);
CREATE INDEX idx_blocks_height ON blocks(chain, height);
CREATE INDEX idx_blocks_hash ON blocks(hash);
*/

-- ============================================
-- 表说明
-- ============================================
-- 1. assets: 存储支持的资产配置信息
-- 2. wallet_accounts: 存储用户的钱包账户信息
-- 3. balances: 存储用户各资产的余额（可用、冻结、待处理）
-- 4. deposits: 存储充值记录，支持按区块范围查询（用于重组回滚）
-- 5. withdrawals: 存储提现请求和状态
-- 6. blocks: 存储区块信息，用于重组检测
--
-- 注意事项：
-- - big.Int 类型使用 NUMERIC(78,0) / DECIMAL(78,0) 存储，支持最大 256 位整数
-- - JSON 字段在 PostgreSQL 使用 JSONB，MySQL 使用 JSON
-- - 所有时间戳字段都有默认值
-- - 索引已优化常用查询场景
-- ============================================

