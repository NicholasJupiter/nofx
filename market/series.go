package market

import "math"

// series.go - 系列数据和技术指标计算函数

// calculatePriceChange 计算价格变化百分比
func calculatePriceChange(priceSeries []float64) float64 {
	if len(priceSeries) < 2 {
		return 0
	}
	current := priceSeries[len(priceSeries)-1]
	previous := priceSeries[0]
	if previous > 0 {
		return ((current - previous) / previous) * 100
	}
	return 0
}

// calculateSimpleATR 简化版ATR计算
func calculateSimpleATR(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	sum := 0.0
	for i := 1; i <= period; i++ {
		tr := math.Abs(prices[i] - prices[i-1])
		sum += tr
	}

	return sum / float64(period)
}

// calculateLongerTermDataFromSeries 从价格序列计算长期数据
func calculateLongerTermDataFromSeries(priceSeries []float64, volume float64) *LongerTermData {
	data := &LongerTermData{
		MACDValues:  make([]float64, 0, 10),
		RSI14Values: make([]float64, 0, 10),
	}

	if len(priceSeries) == 0 {
		return data
	}

	// 从timeframe.go导入的函数计算EMA
	data.EMA20 = calculateEMAFromSeries(priceSeries, 20)
	data.EMA50 = calculateEMAFromSeries(priceSeries, 50)

	// 计算ATR (简化版本，基于价格序列)
	data.ATR14 = calculateSimpleATR(priceSeries, 14)
	data.ATR3 = calculateSimpleATR(priceSeries, 3)

	// 成交量数据
	data.CurrentVolume = volume
	data.AverageVolume = volume // 简化处理

	// 计算MACD和RSI序列
	start := len(priceSeries) - 10
	if start < 0 {
		start = 0
	}

	for i := start; i < len(priceSeries); i++ {
		if i >= 26 {
			macd := calculateMACDFromSeries(priceSeries[:i+1])
			data.MACDValues = append(data.MACDValues, macd)
		}
		if i >= 14 {
			rsi14 := calculateRSIFromSeries(priceSeries[:i+1], 14)
			data.RSI14Values = append(data.RSI14Values, rsi14)
		}
	}

	return data
}

// calculateIntradaySeries 计算日内系列数据
func calculateIntradaySeries(klines []Kline) *IntradayData {
	data := &IntradayData{
		MidPrices:   make([]float64, 0, 10),
		EMA20Values: make([]float64, 0, 10),
		MACDValues:  make([]float64, 0, 10),
		RSI7Values:  make([]float64, 0, 10),
		RSI14Values: make([]float64, 0, 10),
	}

	// 获取最近10个数据点
	start := len(klines) - 10
	if start < 0 {
		start = 0
	}

	for i := start; i < len(klines); i++ {
		data.MidPrices = append(data.MidPrices, klines[i].Close)

		// 计算每个点的EMA20
		if i >= 19 {
			ema20 := calculateEMA(klines[:i+1], 20)
			data.EMA20Values = append(data.EMA20Values, ema20)
		}

		// 计算每个点的MACD
		if i >= 25 {
			macd := calculateMACD(klines[:i+1])
			data.MACDValues = append(data.MACDValues, macd)
		}

		// 计算每个点的RSI
		if i >= 7 {
			rsi7 := calculateRSI(klines[:i+1], 7)
			data.RSI7Values = append(data.RSI7Values, rsi7)
		}
		if i >= 14 {
			rsi14 := calculateRSI(klines[:i+1], 14)
			data.RSI14Values = append(data.RSI14Values, rsi14)
		}
	}

	return data
}

// calculateEMA 计算EMA（基于K线）
func calculateEMA(klines []Kline, period int) float64 {
	if len(klines) < period {
		return 0
	}

	// 计算SMA作为初始EMA
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += klines[i].Close
	}
	ema := sum / float64(period)

	// 计算EMA
	multiplier := 2.0 / float64(period+1)
	for i := period; i < len(klines); i++ {
		ema = (klines[i].Close-ema)*multiplier + ema
	}

	return ema
}

// calculateMACD 计算MACD（基于K线）
func calculateMACD(klines []Kline) float64 {
	if len(klines) < 26 {
		return 0
	}

	// 计算12期和26期EMA
	ema12 := calculateEMA(klines, 12)
	ema26 := calculateEMA(klines, 26)

	// MACD = EMA12 - EMA26
	return ema12 - ema26
}

// calculateRSI 计算RSI（基于K线）
func calculateRSI(klines []Kline, period int) float64 {
	if len(klines) <= period {
		return 0
	}

	gains := 0.0
	losses := 0.0

	// 计算初始平均涨跌幅
	for i := 1; i <= period; i++ {
		change := klines[i].Close - klines[i-1].Close
		if change > 0 {
			gains += change
		} else {
			losses += -change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	// 使用Wilder平滑方法计算后续RSI
	for i := period + 1; i < len(klines); i++ {
		change := klines[i].Close - klines[i-1].Close
		if change > 0 {
			avgGain = (avgGain*float64(period-1) + change) / float64(period)
			avgLoss = (avgLoss * float64(period-1)) / float64(period)
		} else {
			avgGain = (avgGain * float64(period-1)) / float64(period)
			avgLoss = (avgLoss*float64(period-1) + (-change)) / float64(period)
		}
	}

	if avgLoss == 0 {
		return 100
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi
}
