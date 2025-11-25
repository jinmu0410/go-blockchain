#!/bin/bash

# Wallet API 测试脚本
BASE_URL="http://localhost:8081"

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Wallet API 测试 ===${NC}"
echo ""

# 检查服务是否运行
echo -e "${YELLOW}检查服务状态...${NC}"
HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/health")
HTTP_CODE=$(echo "$HEALTH_RESPONSE" | tail -n1)
BODY=$(echo "$HEALTH_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" != "200" ]; then
    echo -e "${YELLOW}错误: 服务响应异常 (HTTP $HTTP_CODE)${NC}"
    echo "响应内容: $BODY"
    echo ""
    echo -e "${YELLOW}请确保服务已启动:${NC}"
    echo "  cd cmd/server && go run main.go"
    exit 1
fi

# 检查是否是 JSON 响应
if echo "$BODY" | grep -q "status"; then
    echo -e "${GREEN}✓ 服务运行正常${NC}"
else
    echo -e "${YELLOW}警告: 服务响应格式异常${NC}"
    echo "响应: $BODY"
fi
echo ""

# 1. 健康检查
echo -e "${BLUE}1. 健康检查${NC}"
if command -v jq &> /dev/null; then
    curl -s "$BASE_URL/health" | jq .
else
    curl -s "$BASE_URL/health"
fi
echo ""

# 2. 注册资产
echo -e "${BLUE}2. 注册资产 (ETH)${NC}"
if command -v jq &> /dev/null; then
    curl -s -X POST "$BASE_URL/api/v1/assets" \
      -H "Content-Type: application/json" \
      -d '{
        "symbol": "ETH",
        "chain": "evm",
        "decimals": 18
      }' | jq .
else
    curl -s -X POST "$BASE_URL/api/v1/assets" \
      -H "Content-Type: application/json" \
      -d '{
        "symbol": "ETH",
        "chain": "evm",
        "decimals": 18
      }'
fi
echo ""

# 3. 创建账户
echo -e "${BLUE}3. 创建账户 (user123, ETH)${NC}"
if command -v jq &> /dev/null; then
    curl -s -X POST "$BASE_URL/api/v1/accounts" \
      -H "Content-Type: application/json" \
      -d '{
        "user_id": "user123",
        "asset_symbol": "ETH"
      }' | jq .
else
    curl -s -X POST "$BASE_URL/api/v1/accounts" \
      -H "Content-Type: application/json" \
      -d '{
        "user_id": "user123",
        "asset_symbol": "ETH"
      }'
fi
echo ""

# 4. 获取账户信息
echo -e "${BLUE}4. 获取账户信息${NC}"
if command -v jq &> /dev/null; then
    curl -s "$BASE_URL/api/v1/accounts/user123/ETH" | jq .
else
    curl -s "$BASE_URL/api/v1/accounts/user123/ETH"
fi
echo ""

# 5. 查询余额
echo -e "${BLUE}5. 查询余额${NC}"
if command -v jq &> /dev/null; then
    curl -s "$BASE_URL/api/v1/balances/user123/ETH" | jq .
else
    curl -s "$BASE_URL/api/v1/balances/user123/ETH"
fi
echo ""

# 6. 列出所有资产
echo -e "${BLUE}6. 列出所有资产${NC}"
if command -v jq &> /dev/null; then
    curl -s "$BASE_URL/api/v1/assets" | jq .
else
    curl -s "$BASE_URL/api/v1/assets"
fi
echo ""

# 7. 列出用户所有余额
echo -e "${BLUE}7. 列出用户所有余额${NC}"
if command -v jq &> /dev/null; then
    curl -s "$BASE_URL/api/v1/balances/user123" | jq .
else
    curl -s "$BASE_URL/api/v1/balances/user123"
fi
echo ""

# 8. 注册 USDT 资产
echo -e "${BLUE}8. 注册资产 (USDT)${NC}"
if command -v jq &> /dev/null; then
    curl -s -X POST "$BASE_URL/api/v1/assets" \
      -H "Content-Type: application/json" \
      -d '{
        "symbol": "USDT",
        "chain": "evm",
        "decimals": 6,
        "token_addr": "0xdAC17F958D2ee523a2206206994597C13D831ec7"
      }' | jq .
else
    curl -s -X POST "$BASE_URL/api/v1/assets" \
      -H "Content-Type: application/json" \
      -d '{
        "symbol": "USDT",
        "chain": "evm",
        "decimals": 6,
        "token_addr": "0xdAC17F958D2ee523a2206206994597C13D831ec7"
      }'
fi
echo ""

echo -e "${GREEN}=== 测试完成 ===${NC}"

