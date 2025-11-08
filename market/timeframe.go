package market

import (
	"fmt"
	"log"
	"math"
)

// ==================== 多时间框架分析函数 ====================

// getMultiTimeframeData 获取多时间框架数据（根据配置动态获取）
func getMultiTimeframeData(symbol string) (MultiTimeframeData, error) {
	if WSMonitorCli == nil || len(WSMonitorCli.intervals) == 0 {
		return nil, fmt.Errorf("WSMonitor未初始化或周期配置为空")
	}

	multiTimeframe := make(MultiTimeframeData)

	// 遍历所有配置的时间周期
	for _, interval := range WSMonitorCli.intervals {
		klines, err := WSMonitorCli.GetCurrentKlines(symbol, interval)
		if err != nil {
			// 单个周期失败不影响其他周期
			log.Printf("⚠️  获取 %s %s K线失败: %v", symbol, interval, err)
			continue
		}

		// 计算该时间框架的数据
		timeframeData := calculateTimeframeData(klines, interval)
		multiTimeframe[interval] = timeframeData
	}

	if len(multiTimeframe) == 0 {
		return nil, fmt.Errorf("无法获取任何时间框架的数据")
	}

	return multiTimeframe, nil
}

// calculateTimeframeData 计算单个时间框架的技术指标
func calculateTimeframeData(klines []Kline, timeframe string) *TimeframeData {
	if len(klines) == 0 {
		return &TimeframeData{Timeframe: timeframe}
	}

	currentPrice := klines[len(klines)-1].Close

	// 提取价格序列
	priceSeries := make([]float64, len(klines))
	for i, k := range klines {
		priceSeries[i] = k.Close
	}

	// 计算技术指标
	ema20 := calculateEMAFromSeries(priceSeries, 20)
	ema50 := calculateEMAFromSeries(priceSeries, 50)
	macd := calculateMACDFromSeries(priceSeries)
	rsi7 := calculateRSIFromSeries(priceSeries, 7)
	rsi14 := calculateRSIFromSeries(priceSeries, 14)
	atr14 := calculateATRFromKlines(klines, 14)

	volume := 0.0
	if len(klines) > 0 {
		volume = klines[len(klines)-1].Volume
	}

	// 判断趋势方向
	trendDirection := determineTrendDirection(currentPrice, ema20, ema50, macd)

	// 计算信号强度
	signalStrength := calculateSignalStrength(currentPrice, ema20, ema50, macd, rsi7)

	return &TimeframeData{
		Timeframe:      timeframe,
		CurrentPrice:   currentPrice,
		EMA20:          ema20,
		EMA50:          ema50,
		MACD:           macd,
		RSI7:           rsi7,
		RSI14:          rsi14,
		ATR14:          atr14,
		Volume:         volume,
		Klines:         klines,
		PriceSeries:    priceSeries,
		TrendDirection: trendDirection,
		SignalStrength: signalStrength,
	}
}

// calculateEMAFromSeries 基于价格序列计算EMA
func calculateEMAFromSeries(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	// 计算SMA作为初始EMA
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	ema := sum / float64(period)

	// 计算EMA
	multiplier := 2.0 / float64(period+1)
	for i := period; i < len(prices); i++ {
		ema = (prices[i]-ema)*multiplier + ema
	}

	return ema
}

// calculateMACDFromSeries 基于价格序列计算MACD
func calculateMACDFromSeries(prices []float64) float64 {
	if len(prices) < 26 {
		return 0
	}

	ema12 := calculateEMAFromSeries(prices, 12)
	ema26 := calculateEMAFromSeries(prices, 26)

	return ema12 - ema26
}

// calculateRSIFromSeries 基于价格序列计算RSI
func calculateRSIFromSeries(prices []float64, period int) float64 {
	if len(prices) <= period {
		return 0
	}

	gains := 0.0
	losses := 0.0

	for i := 1; i <= period; i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains += change
		} else {
			losses += -change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	if avgLoss == 0 {
		return 100
	}

	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}

// calculateATRFromKlines 基于K线计算ATR
func calculateATRFromKlines(klines []Kline, period int) float64 {
	if len(klines) <= period {
		return 0
	}

	trs := make([]float64, len(klines))
	for i := 1; i < len(klines); i++ {
		high := klines[i].High
		low := klines[i].Low
		prevClose := klines[i-1].Close

		tr1 := high - low
		tr2 := math.Abs(high - prevClose)
		tr3 := math.Abs(low - prevClose)

		trs[i] = math.Max(tr1, math.Max(tr2, tr3))
	}

	sum := 0.0
	for i := 1; i <= period; i++ {
		sum += trs[i]
	}
	atr := sum / float64(period)

	for i := period + 1; i < len(klines); i++ {
		atr = (atr*float64(period-1) + trs[i]) / float64(period)
	}

	return atr
}

// determineTrendDirection 判断趋势方向
func determineTrendDirection(price, ema20, ema50, macd float64) string {
	bullishSignals := 0
	bearishSignals := 0

	// 信号1: 价格与 EMA20 关系
	if price > ema20 && ema20 > 0 {
		bullishSignals++
	} else if price < ema20 && ema20 > 0 {
		bearishSignals++
	}

	// 信号2: EMA20 与 EMA50 关系（趋势方向）
	if ema20 > ema50 && ema50 > 0 {
		bullishSignals++
	} else if ema20 < ema50 && ema50 > 0 {
		bearishSignals++
	}

	// 信号3: MACD 趋势
	// ⚠️ 方案A修复：动态计算 MACD 阈值（相对于价格）
	// 原配置（注释）：固定阈值 0.001，对高价币种（如 BTC ~100,000）太小
	// if macd > 0.001 {
	// 	bullishSignals++
	// } else if macd < -0.001 {
	// 	bearishSignals++
	// }

	// 新配置：使用价格的 0.01% 作为动态阈值
	// 例如：BTC (100,000) 阈值 = 10，ETH (3,500) 阈值 = 0.35
	macdThreshold := price * 0.0001 // 0.01% 的价格
	if macd > macdThreshold {
		bullishSignals++
	} else if macd < -macdThreshold {
		bearishSignals++
	}

	// ⚠️ 方案A修复：提高趋势判断标准，要求 3 个信号都满足（而不是 2 个）
	// 原配置（注释）：只需要 2/3 信号满足，容易产生虚假信号
	// if bullishSignals >= 2 {
	// 	return "bullish"
	// } else if bearishSignals >= 2 {
	// 	return "bearish"
	// }

	// 新配置：需要 3 个信号都满足，确保趋势明确
	if bullishSignals == 3 {
		return "bullish"
	} else if bearishSignals == 3 {
		return "bearish"
	}
	return "neutral"
}

// calculateSignalStrength 计算信号强度
func calculateSignalStrength(price, ema20, ema50, macd, rsi7 float64) int {
	strength := 50

	// 价格与EMA关系
	if price > ema20 && ema20 > ema50 {
		strength += 20
	} else if price < ema20 && ema20 < ema50 {
		strength -= 20
	}

	// MACD信号
	if macd > 0.001 {
		strength += 15
	} else if macd < -0.001 {
		strength -= 15
	}

	// RSI信号
	if rsi7 < 30 {
		strength += 10
	} else if rsi7 > 70 {
		strength -= 10
	}

	if strength < 0 {
		return 0
	}
	if strength > 100 {
		return 100
	}
	return strength
}
