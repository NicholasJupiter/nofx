package market

// ==================== 斐波那契和市场结构计算函数 ====================

// calculateFibonacciLevels 计算斐波那契回撤水平
func calculateFibonacciLevels(high, low float64) *FibLevels {
	diff := high - low
	return &FibLevels{
		Level236: high - (diff * 0.236),
		Level382: high - (diff * 0.382),
		Level500: high - (diff * 0.5),
		Level618: high - (diff * 0.618),
		Level705: high - (diff * 0.705),
		Level786: high - (diff * 0.786),
		High:     high,
		Low:      low,
		Trend:    "bullish", // 默认，实际使用时需要根据趋势判断
	}
}

// detectMarketStructure 检测市场结构
func detectMarketStructure(priceSeries []float64) *MarketStructure {
	if len(priceSeries) < 10 {
		return nil
	}

	structure := &MarketStructure{
		SwingHighs: make([]float64, 0),
		SwingLows:  make([]float64, 0),
	}

	// 简单的波段检测算法
	for i := 2; i < len(priceSeries)-2; i++ {
		// 检测波段高点
		if priceSeries[i] > priceSeries[i-1] && priceSeries[i] > priceSeries[i-2] &&
			priceSeries[i] > priceSeries[i+1] && priceSeries[i] > priceSeries[i+2] {
			structure.SwingHighs = append(structure.SwingHighs, priceSeries[i])
		}
		// 检测波段低点
		if priceSeries[i] < priceSeries[i-1] && priceSeries[i] < priceSeries[i-2] &&
			priceSeries[i] < priceSeries[i+1] && priceSeries[i] < priceSeries[i+2] {
			structure.SwingLows = append(structure.SwingLows, priceSeries[i])
		}
	}

	// 确定当前偏向
	if len(structure.SwingHighs) > 1 && len(structure.SwingLows) > 1 {
		latestHigh := structure.SwingHighs[len(structure.SwingHighs)-1]
		prevHigh := structure.SwingHighs[len(structure.SwingHighs)-2]
		latestLow := structure.SwingLows[len(structure.SwingLows)-1]
		prevLow := structure.SwingLows[len(structure.SwingLows)-2]

		if latestHigh > prevHigh && latestLow > prevLow {
			structure.CurrentBias = "bullish"
		} else if latestHigh < prevHigh && latestLow < prevLow {
			structure.CurrentBias = "bearish"
		} else {
			structure.CurrentBias = "neutral"
		}
	}

	return structure
}

// calculateCurrentFibLevels 计算当前斐波那契水平
func calculateCurrentFibLevels(structure *MarketStructure) *FibLevels {
	if structure == nil || len(structure.SwingHighs) < 2 || len(structure.SwingLows) < 2 {
		return nil
	}

	// 使用最近的波段高点和低点
	recentHigh := structure.SwingHighs[len(structure.SwingHighs)-1]
	recentLow := structure.SwingLows[len(structure.SwingLows)-1]

	// 确保高点高于低点
	if recentHigh <= recentLow {
		return nil
	}

	fibLevels := calculateFibonacciLevels(recentHigh, recentLow)
	fibLevels.Trend = structure.CurrentBias

	return fibLevels
}

