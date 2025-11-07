package market

import "math"

// condition.go - 市场状态检测函数

// MarketCondition 市场状态结构
type MarketCondition struct {
	Condition    string  // "trending", "ranging", "volatile"
	Confidence   int     // 0-100
	ATRRatio     float64 // ATR/Price 比率
	EMASlope     float64 // EMA20斜率
	PriceChannel float64 // 价格通道宽度
}

// DetectMarketCondition 检测市场状态
func DetectMarketCondition(data *Data) *MarketCondition {
	if data == nil {
		return &MarketCondition{Condition: "unknown", Confidence: 0}
	}

	condition := &MarketCondition{}

	// 使用现有数据计算市场状态
	atrRatio := calculateATRRatio(data)
	emaSlope := calculateEMASlope(data)
	priceChannel := calculatePriceChannel(data)
	rsiPosition := analyzeRSIPosition(data)
	timeframeConsistency := checkTimeframeConsistency(data)

	trendingScore, rangingScore := calculateMarketScores(
		atrRatio, emaSlope, priceChannel, rsiPosition, timeframeConsistency)

	if trendingScore > 70 {
		condition.Condition = "trending"
		condition.Confidence = trendingScore
	} else if rangingScore > 60 {
		condition.Condition = "ranging"
		condition.Confidence = rangingScore
	} else {
		condition.Condition = "volatile"
		condition.Confidence = 50
	}

	condition.ATRRatio = atrRatio
	condition.EMASlope = emaSlope
	condition.PriceChannel = priceChannel

	return condition
}

// calculateATRRatio 基于现有ATR数据计算波动率
func calculateATRRatio(data *Data) float64 {
	if data.LongerTermContext == nil || data.CurrentPrice == 0 {
		return 0
	}
	return (data.LongerTermContext.ATR14 / data.CurrentPrice) * 100
}

// calculateEMASlope 基于现有EMA数据计算斜率
func calculateEMASlope(data *Data) float64 {
	// 方法1：使用多时间框架EMA值估算斜率
	if data.MultiTimeframe != nil {
		var emaValues []float64

		// 遍历所有配置的时间框架收集 EMA20 值
		if WSMonitorCli != nil && len(WSMonitorCli.intervals) > 0 {
			for _, interval := range WSMonitorCli.intervals {
				if tf, ok := (*data.MultiTimeframe)[interval]; ok && tf != nil {
					emaValues = append(emaValues, tf.EMA20)
				}
			}
		}

		if len(emaValues) >= 2 {
			// 计算EMA变化的百分比斜率
			slope := (emaValues[len(emaValues)-1] - emaValues[0]) / emaValues[0] * 100
			return slope
		}
	}

	// 方法2：使用当前EMA和历史EMA（如果有）
	if data.LongerTermContext != nil && data.LongerTermContext.EMA20 != 0 {
		slope := (data.CurrentEMA20 - data.LongerTermContext.EMA20) / data.LongerTermContext.EMA20 * 100
		return slope
	}

	return 0
}

// calculatePriceChannel 计算价格通道宽度
func calculatePriceChannel(data *Data) float64 {
	// 使用多时间框架的最高最低EMA估算通道
	if data.MultiTimeframe == nil {
		return 0
	}

	var emas []float64

	// 遍历所有配置的时间框架收集 EMA20 值
	if WSMonitorCli != nil && len(WSMonitorCli.intervals) > 0 {
		for _, interval := range WSMonitorCli.intervals {
			if tf, ok := (*data.MultiTimeframe)[interval]; ok && tf != nil {
				emas = append(emas, tf.EMA20)
			}
		}
	}

	if len(emas) < 2 {
		return 0
	}

	// 找到EMA的最大最小值
	minEMA, maxEMA := emas[0], emas[0]
	for _, ema := range emas {
		if ema < minEMA {
			minEMA = ema
		}
		if ema > maxEMA {
			maxEMA = ema
		}
	}

	channelWidth := (maxEMA - minEMA) / data.CurrentPrice * 100
	return channelWidth
}

// analyzeRSIPosition 分析RSI位置
func analyzeRSIPosition(data *Data) float64 {
	// 使用现有RSI数据判断是否在震荡区间
	rsiValue := data.CurrentRSI7

	// 判断RSI是否在震荡区间 (30-70)
	if rsiValue >= 30 && rsiValue <= 70 {
		return 80 // 高概率震荡
	} else if rsiValue >= 40 && rsiValue <= 60 {
		return 95 // 极高概率震荡
	} else {
		return 30 // 低概率震荡
	}
}

// checkTimeframeConsistency 检查多时间框架一致性
func checkTimeframeConsistency(data *Data) float64 {
	if data.MultiTimeframe == nil {
		return 0
	}

	bullishCount, bearishCount := 0, 0
	validCount := 0

	// 遍历所有配置的时间框架
	if WSMonitorCli != nil && len(WSMonitorCli.intervals) > 0 {
		for _, interval := range WSMonitorCli.intervals {
			if tf, ok := (*data.MultiTimeframe)[interval]; ok && tf != nil {
				validCount++
				if tf.TrendDirection == "bullish" {
					bullishCount++
				} else if tf.TrendDirection == "bearish" {
					bearishCount++
				}
			}
		}
	}

	if validCount == 0 {
		return 0
	}

	// 计算一致性得分
	consistency := math.Max(float64(bullishCount), float64(bearishCount)) / float64(validCount) * 100
	return consistency
}

// calculateMarketScores 计算市场状态得分
func calculateMarketScores(atrRatio, emaSlope, priceChannel, rsiPosition, timeframeConsistency float64) (int, int) {
	trendingScore, rangingScore := 0, 0

	// 趋势市特征
	if math.Abs(emaSlope) > 0.1 { // EMA有明显斜率
		trendingScore += 25
	}
	if atrRatio > 0.3 { // 波动率适中偏高
		trendingScore += 20
	}
	if timeframeConsistency > 70 { // 多时间框架一致
		trendingScore += 30
	}
	if rsiPosition < 50 { // RSI不在中间区域
		trendingScore += 25
	}

	// 震荡市特征
	if math.Abs(emaSlope) < 0.05 { // EMA走平
		rangingScore += 30
	}
	if priceChannel < 2.0 { // 价格通道狭窄
		rangingScore += 25
	}
	if rsiPosition > 70 { // RSI常在中间区域
		rangingScore += 25
	}
	if timeframeConsistency < 50 { // 多时间框架不一致
		rangingScore += 20
	}

	return trendingScore, rangingScore
}

// IsRangingMarket 判断是否为震荡市
func IsRangingMarket(data *Data) bool {
	condition := DetectMarketCondition(data)
	return condition.Condition == "ranging" && condition.Confidence > 60
}
