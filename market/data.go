package market

import (
	"fmt"
)

// Get 获取指定代币的市场数据
func Get(symbol string) (*Data, error) {
	// 标准化symbol
	symbol = Normalize(symbol)

	// 获取多时间框架数据
	multiTimeframe, err := getMultiTimeframeData(symbol)
	if err != nil {
		return nil, fmt.Errorf("获取多时间框架数据失败: %v", err)
	}

	// 使用 3m 作为主要参考时间框架
	primaryData, ok := multiTimeframe["15m"]
	if !ok {
		return nil, fmt.Errorf("未找到 15m 时间框架数据")
	}

	var priceChange1h, priceChange4h, priceChange1d float64

	if tf, ok := multiTimeframe["1h"]; ok {
		priceChange1d = calculatePriceChange(tf.PriceSeries)
	}
	if tf, ok := multiTimeframe["4h"]; ok {
		priceChange4h = calculatePriceChange(tf.PriceSeries)
	}
	if tf, ok := multiTimeframe["1d"]; ok {
		priceChange1d = calculatePriceChange(tf.PriceSeries)
	}

	// 获取OI数据
	oiData, err := getOpenInterestData(symbol)
	if err != nil {
		oiData = &OIData{Latest: 0, Average: 0}
	}

	// 获取Funding Rate
	fundingRate, _ := getFundingRate(symbol)

	// 计算长期数据（基于 4h）
	var longerTermData *LongerTermData
	if tf4h, ok := multiTimeframe["4h"]; ok {
		longerTermData = calculateLongerTermDataFromSeries(tf4h.PriceSeries, tf4h.Volume)
	}

	// 计算日内系列数据（基于 15m）
	intradayData := calculateIntradaySeries(primaryData.Klines)

	// 计算市场结构和斐波那契水平（使用 1d）
	var marketStructure *MarketStructure
	var fibLevels *FibLevels

	if tf1d, ok := multiTimeframe["1d"]; ok {
		marketStructure = detectMarketStructure(tf1d.PriceSeries)
		if marketStructure != nil {
			fibLevels = calculateCurrentFibLevels(marketStructure)
			marketStructure.FibLevels = fibLevels
		}
	}

	return &Data{
		Symbol:            symbol,
		CurrentPrice:      primaryData.CurrentPrice,
		PriceChange1h:     priceChange1h,
		PriceChange4h:     priceChange4h,
		PriceChange1d:     priceChange1d, // 新增：日线价格变化
		CurrentEMA20:      primaryData.EMA20,
		CurrentMACD:       primaryData.MACD,
		CurrentRSI7:       primaryData.RSI7,
		OpenInterest:      oiData,
		FundingRate:       fundingRate,
		IntradaySeries:    intradayData,
		LongerTermContext: longerTermData,
		MultiTimeframe:    &multiTimeframe,
		MarketStructure:   marketStructure,
		FibLevels:         fibLevels,
	}, nil
}
