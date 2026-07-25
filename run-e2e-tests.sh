#!/bin/bash
# run-e2e-tests.sh
# 运行端到端真实测试的脚本
# 用法:
#   1. 在浏览器中打开 https://sploitus.com
#   2. F12 > Application > Cookies > 复制所有 Cookie 字符串
#   3. 运行: SPLOITUS_COOKIE="cookie_string" ./run-e2e-tests.sh
#   4. （可选）设置代理: SPLOITUS_PROXY="http://localhost:8080" ./run-e2e-tests.sh

set -e

if [ -z "$SPLOITUS_COOKIE" ]; then
    echo "❌ 请先设置 SPLOITUS_COOKIE 环境变量"
    echo ""
    echo "使用方式:"
    echo "  1. 在 Chrome 中打开 https://sploitus.com"
    echo "  2. 按 F12 打开开发者工具"
    echo "  3. 进入 Application > Cookies > sploitus.com"
    echo "  4. 右键点击任意 Cookie > 复制所有 Cookie"
    echo "  5. 运行:"
    echo "     SPLOITUS_COOKIE='_ym_uid=...; cf_clearance=...' ./run-e2e-tests.sh"
    exit 1
fi

echo "🚀 运行端到端真实测试..."
echo "Cookie 长度: ${#SPLOITUS_COOKIE} 字符"
echo ""

export SPLOITUS_COOKIE

go test ./pkg/sploitus -run "TestE2E" -v -count=1 2>&1 | grep -E "PASS|FAIL|SKIP|---"