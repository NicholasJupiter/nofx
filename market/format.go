package market

import (
	"fmt"
	"strings"
)

// format.go - 数据格式化输出函数

// Format 格式化输出市场数据
func Format(data *Data) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("current_price = %.2f, current_ema20 = %.3f, current_macd = %.3f, current_rsi (7 period) = %.3f\n\n",
		data.CurrentPrice, data.CurrentEMA20, data.CurrentMACD, data.CurrentRSI7))

	sb.WriteString(fmt.Sprintf("In addition, here is the latest %s open interest and funding rate for perps:\n\n",
		data.Symbol))

	if data.OpenInterest != nil {
		sb.WriteString(fmt.Sprintf("Open Interest: Latest: %.2f Average: %.2f\n\n",
			data.OpenInterest.Latest, data.OpenInterest.Average))
	}

	sb.WriteString(fmt.Sprintf("Funding Rate: %.2e\n\n", data.FundingRate))

	if data.IntradaySeries != nil {
		sb.WriteString("Intraday series (3‑minute intervals, oldest → latest):\n\n")

		if len(data.IntradaySeries.MidPrices) > 0 {
			sb.WriteString(fmt.Sprintf("Mid prices: %s\n\n", formatFloatSlice(data.IntradaySeries.MidPrices)))
		}

		if len(data.IntradaySeries.EMA20Values) > 0 {
			sb.WriteString(fmt.Sprintf("EMA indicators (20‑period): %s\n\n", formatFloatSlice(data.IntradaySeries.EMA20Values)))
		}

		if len(data.IntradaySeries.MACDValues) > 0 {
			sb.WriteString(fmt.Sprintf("MACD indicators: %s\n\n", formatFloatSlice(data.IntradaySeries.MACDValues)))
		}

		if len(data.IntradaySeries.RSI7Values) > 0 {
			sb.WriteString(fmt.Sprintf("RSI indicators (7‑Period): %s\n\n", formatFloatSlice(data.IntradaySeries.RSI7Values)))
		}

		if len(data.IntradaySeries.RSI14Values) > 0 {
			sb.WriteString(fmt.Sprintf("RSI indicators (14‑Period): %s\n\n", formatFloatSlice(data.IntradaySeries.RSI14Values)))
		}
	}

	if data.LongerTermContext != nil {
		sb.WriteString("Longer‑term context (4‑hour timeframe):\n\n")

		sb.WriteString(fmt.Sprintf("20‑Period EMA: %.3f vs. 50‑Period EMA: %.3f\n\n",
			data.LongerTermContext.EMA20, data.LongerTermContext.EMA50))

		sb.WriteString(fmt.Sprintf("3‑Period ATR: %.3f vs. 14‑Period ATR: %.3f\n\n",
			data.LongerTermContext.ATR3, data.LongerTermContext.ATR14))

		sb.WriteString(fmt.Sprintf("Current Volume: %.3f vs. Average Volume: %.3f\n\n",
			data.LongerTermContext.CurrentVolume, data.LongerTermContext.AverageVolume))

		if len(data.LongerTermContext.MACDValues) > 0 {
			sb.WriteString(fmt.Sprintf("MACD indicators: %s\n\n", formatFloatSlice(data.LongerTermContext.MACDValues)))
		}

		if len(data.LongerTermContext.RSI14Values) > 0 {
			sb.WriteString(fmt.Sprintf("RSI indicators (14‑Period): %s\n\n", formatFloatSlice(data.LongerTermContext.RSI14Values)))
		}
	}

	// 多时间框架信息
	if data.MultiTimeframe != nil && len(*data.MultiTimeframe) > 0 {
		sb.WriteString("Multi‑Timeframe Analysis:\n\n")

		// 遍历所有配置的时间框架（按配置顺序）
		if WSMonitorCli != nil && len(WSMonitorCli.intervals) > 0 {
			for _, interval := range WSMonitorCli.intervals {
				if tf, ok := (*data.MultiTimeframe)[interval]; ok && tf != nil {
					sb.WriteString(fmt.Sprintf("%s Timeframe: %s (Strength: %d/100)\n",
						interval, tf.TrendDirection, tf.SignalStrength))
					sb.WriteString(fmt.Sprintf("  Price: %.4f | EMA20: %.4f | EMA50: %.4f\n",
						tf.CurrentPrice, tf.EMA20, tf.EMA50))
					sb.WriteString(fmt.Sprintf("  MACD: %.4f | RSI7: %.2f | RSI14: %.2f | ATR14: %.4f\n\n",
						tf.MACD, tf.RSI7, tf.RSI14, tf.ATR14))
				}
			}
		}
	}

	// 市场结构信息
	if data.MarketStructure != nil {
		sb.WriteString("Market Structure:\n\n")
		sb.WriteString(fmt.Sprintf("Current Bias: %s\n\n", data.MarketStructure.CurrentBias))
		sb.WriteString(fmt.Sprintf("Swing Highs detected: %d | Swing Lows detected: %d\n\n",
			len(data.MarketStructure.SwingHighs), len(data.MarketStructure.SwingLows)))

		if len(data.MarketStructure.SwingHighs) > 0 && len(data.MarketStructure.SwingLows) > 0 {
			recentHigh := data.MarketStructure.SwingHighs[len(data.MarketStructure.SwingHighs)-1]
			recentLow := data.MarketStructure.SwingLows[len(data.MarketStructure.SwingLows)-1]
			sb.WriteString(fmt.Sprintf("Recent Swing: High %.4f → Low %.4f\n\n", recentHigh, recentLow))
		}
	}

	// 斐波那契水平信息
	if data.FibLevels != nil {
		sb.WriteString("Fibonacci Levels:\n\n")
		sb.WriteString(fmt.Sprintf("Trend: %s | Range: %.4f (High) - %.4f (Low)\n\n",
			data.FibLevels.Trend, data.FibLevels.High, data.FibLevels.Low))
		sb.WriteString(fmt.Sprintf("  0.236: %.4f | 0.382: %.4f | 0.500: %.4f\n\n",
			data.FibLevels.Level236, data.FibLevels.Level382, data.FibLevels.Level500))
		sb.WriteString(fmt.Sprintf("  0.618: %.4f | 0.705: %.4f | 0.786: %.4f\n\n",
			data.FibLevels.Level618, data.FibLevels.Level705, data.FibLevels.Level786))

		// OTE区间（0.618-0.705）
		sb.WriteString(fmt.Sprintf("OTE Zone (Optimal Trade Entry): %.4f - %.4f\n\n",
			data.FibLevels.Level705, data.FibLevels.Level618))

		// 判断当前价格位置
		currentPrice := data.CurrentPrice
		if currentPrice >= data.FibLevels.Level705 && currentPrice <= data.FibLevels.Level618 {
			sb.WriteString("**Current Price in OTE Golden Zone**\n\n")
		} else if currentPrice > data.FibLevels.Level500 {
			sb.WriteString("Current Price in Premium Zone\n\n")
		} else {
			sb.WriteString("Current Price in Discount Zone\n\n")
		}
	}

	return sb.String()
}

// formatFloatSlice 格式化float64切片为字符串
func formatFloatSlice(values []float64) string {
	strValues := make([]string, len(values))
	for i, v := range values {
		strValues[i] = fmt.Sprintf("%.3f", v)
	}
	return "[" + strings.Join(strValues, ", ") + "]"
}
