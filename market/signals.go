package market

// signals.go - 交易信号计算函数

// TrendSignal 趋势信号结构
type TrendSignal struct {
	EMACrossover       bool   // EMA20上穿EMA50
	EMA50SlopePositive bool   // EMA50斜率为正
	TrendConfirmed     bool   // 趋势确认（两个条件都满足）
	HTFDirection       string // 高时间框架趋势方向 "bullish", "bearish", "neutral"
	LTFDirection       string // 低时间框架趋势方向 "bullish", "bearish", "neutral"
	HTFLTFAligned      bool   // 高低时间框架是否一致
	CanLong            bool   // 是否允许做多（HTF看多且LTF看多）
	CanShort           bool   // 是否允许做空（HTF看空且LTF看空）
}

// CalculateTrendSignals 计算趋势确认信号
func CalculateTrendSignals(data *Data) *TrendSignal {
	if data == nil || data.MultiTimeframe == nil {
		return &TrendSignal{}
	}

	signal := &TrendSignal{}

	// 1. 计算 EMA 多周期确认（使用主要时间框架，如 3m）
	if primaryTF, ok := (*data.MultiTimeframe)["3m"]; ok && primaryTF != nil {
		signal.EMACrossover = checkEMACrossover(primaryTF)
		signal.EMA50SlopePositive = checkEMA50Slope(primaryTF)
		signal.TrendConfirmed = signal.EMACrossover && signal.EMA50SlopePositive
	}

	// 2. 计算 Higher Timeframe Filter（高时间框架过滤）
	// 使用 4h 作为高时间框架，3m 作为低时间框架
	htfTF, htfOk := (*data.MultiTimeframe)["4h"]
	ltfTF, ltfOk := (*data.MultiTimeframe)["3m"]

	if htfOk && ltfOk && htfTF != nil && ltfTF != nil {
		signal.HTFDirection = htfTF.TrendDirection
		signal.LTFDirection = ltfTF.TrendDirection

		// 检查高低时间框架是否一致
		signal.HTFLTFAligned = (signal.HTFDirection == signal.LTFDirection) &&
			(signal.HTFDirection == "bullish" || signal.HTFDirection == "bearish")

		// 判断是否允许做多/做空
		signal.CanLong = signal.HTFDirection == "bullish" && signal.LTFDirection == "bullish"
		signal.CanShort = signal.HTFDirection == "bearish" && signal.LTFDirection == "bearish"
	}

	return signal
}

// checkEMACrossover 检测 EMA20 是否上穿 EMA50
// 判断标准：当前 EMA20 > EMA50，且 EMA20 接近 EMA50（刚上穿不久）
func checkEMACrossover(tf *TimeframeData) bool {
	if tf == nil || tf.EMA20 == 0 || tf.EMA50 == 0 {
		return false
	}

	// EMA20 必须在 EMA50 之上
	if tf.EMA20 <= tf.EMA50 {
		return false
	}

	// 计算 EMA20 和 EMA50 的差距百分比
	// 如果差距小于 1%，说明刚上穿或正在上穿
	gap := (tf.EMA20 - tf.EMA50) / tf.EMA50 * 100

	// 差距在 0% 到 2% 之间，说明是有效的上穿信号
	return gap > 0 && gap < 2.0
}

// checkEMA50Slope 检查 EMA50 斜率是否为正
func checkEMA50Slope(tf *TimeframeData) bool {
	if tf == nil || len(tf.PriceSeries) < 2 {
		return false
	}

	// 使用价格序列重新计算最近的 EMA50 值来判断斜率
	// 计算倒数第2根K线的 EMA50
	if len(tf.PriceSeries) < 51 {
		return false
	}

	// 计算前一个周期的 EMA50
	previousEMA50 := calculateEMAFromSeries(tf.PriceSeries[:len(tf.PriceSeries)-1], 50)
	currentEMA50 := tf.EMA50

	// 如果当前 EMA50 大于前一个周期的 EMA50，说明斜率为正
	return currentEMA50 > previousEMA50
}

// CheckMultiTimeframeAlignment 检查多时间框架趋势一致性（通用版本）
func CheckMultiTimeframeAlignment(data *Data, htfInterval, ltfInterval string) (bool, string, string) {
	if data == nil || data.MultiTimeframe == nil {
		return false, "neutral", "neutral"
	}

	htfTF, htfOk := (*data.MultiTimeframe)[htfInterval]
	ltfTF, ltfOk := (*data.MultiTimeframe)[ltfInterval]

	if !htfOk || !ltfOk || htfTF == nil || ltfTF == nil {
		return false, "neutral", "neutral"
	}

	htfDirection := htfTF.TrendDirection
	ltfDirection := ltfTF.TrendDirection

	// 检查是否一致
	aligned := (htfDirection == ltfDirection) &&
		(htfDirection == "bullish" || htfDirection == "bearish")

	return aligned, htfDirection, ltfDirection
}

// GetTradingPermission 获取交易许可（基于多时间框架过滤）
func GetTradingPermission(data *Data) (canLong bool, canShort bool, reason string) {
	signal := CalculateTrendSignals(data)

	if signal.CanLong {
		return true, false, "HTF(4h) 和 LTF(3m) 均看多，允许做多"
	}

	if signal.CanShort {
		return false, true, "HTF(4h) 和 LTF(3m) 均看空，允许做空"
	}

	// 分析不允许交易的原因
	if signal.HTFDirection == "neutral" || signal.LTFDirection == "neutral" {
		return false, false, "高时间框架或低时间框架趋势不明确"
	}

	if signal.HTFDirection != signal.LTFDirection {
		return false, false, "高低时间框架趋势不一致，禁止开仓"
	}

	return false, false, "未满足交易条件"
}

// IsStrongTrendConfirmed 判断是否有强趋势确认
func IsStrongTrendConfirmed(data *Data) bool {
	signal := CalculateTrendSignals(data)

	// 需要同时满足：
	// 1. EMA交叉确认
	// 2. EMA50斜率为正
	// 3. 高低时间框架一致
	return signal.TrendConfirmed && signal.HTFLTFAligned
}
