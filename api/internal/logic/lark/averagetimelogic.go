package lark

import (
	"context"
	"errors"
	"time"

	"apitools/api/internal/consts"
	"apitools/api/internal/svc"
	"apitools/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AverageTimeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAverageTimeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AverageTimeLogic {
	return &AverageTimeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AverageTimeLogic) AverageTime(req *types.AverageTimeReq) (resp *types.AverageTimeResp, err error) {
	switch req.CalculateType {
	case consts.HourMinuteSecond:
		return l.AverageTimeHourMinuteSecond(req)
	default:
		return nil, errors.New("calculate type not supported")
	}
}

func (l *AverageTimeLogic) AverageTimeHourMinuteSecond(req *types.AverageTimeReq) (resp *types.AverageTimeResp, err error) {
	// 验证输入参数
	if len(req.TimestampList) == 0 {
		return nil, errors.New("timestamp list cannot be empty")
	}

	// 解析时间戳并提取时间单位
	timeUnits, err := l.parseTimestamps(req.TimestampList)
	if err != nil {
		return nil, err
	}

	// 计算平均时间
	averageTime, err := l.calculateAverageTime(timeUnits)
	if err != nil {
		return nil, err
	}

	return &types.AverageTimeResp{
		AverageTime:      time.Unix(averageTime, 0).Format("2006-01-02 15:04:05"),
		AverageTimestamp: averageTime,
		HHMMSS:           time.Unix(averageTime, 0).Format("15:04:05"),
	}, nil
}

// parseTimestamps 解析时间戳列表，提取时间单位
func (l *AverageTimeLogic) parseTimestamps(timestamps []int64) ([]TimeUnit, error) {
	var timeUnits []TimeUnit

	for _, ts := range timestamps {
		// 将时间戳转换为时间对象
		t := time.Unix(ts, 0)

		// 提取时间单位
		timeUnit := TimeUnit{
			Hour:   t.Hour(),
			Minute: t.Minute(),
			Second: t.Second(),
		}

		timeUnits = append(timeUnits, timeUnit)
	}

	return timeUnits, nil
}

// calculateAverageTime 计算平均时间
func (l *AverageTimeLogic) calculateAverageTime(timeUnits []TimeUnit) (int64, error) {
	if len(timeUnits) == 0 {
		return 0, errors.New("no time units to calculate")
	}

	// 确定时间精度（检查是否有秒级精度）
	hasSecondPrecision := l.hasSecondPrecision(timeUnits)

	// 计算总秒数
	var totalSeconds int64
	for _, tu := range timeUnits {
		seconds := int64(tu.Hour)*3600 + int64(tu.Minute)*60 + int64(tu.Second)
		totalSeconds += seconds
	}

	// 计算平均秒数
	avgSeconds := totalSeconds / int64(len(timeUnits))

	// 根据精度决定是否保留秒
	if !hasSecondPrecision {
		// 如果没有秒级精度，将秒数四舍五入到分钟
		avgSeconds = (avgSeconds + 30) / 60 * 60
	}

	// 转换为时间戳（使用今天的日期）
	now := time.Now()
	avgTime := time.Date(now.Year(), now.Month(), now.Day(),
		int(avgSeconds/3600),
		int((avgSeconds%3600)/60),
		int(avgSeconds%60),
		0, now.Location())

	return avgTime.Unix(), nil
}

// hasSecondPrecision 检查时间戳列表是否具有秒级精度
func (l *AverageTimeLogic) hasSecondPrecision(timeUnits []TimeUnit) bool {
	// 如果所有时间戳的秒数都是0，则认为没有秒级精度
	for _, tu := range timeUnits {
		if tu.Second != 0 {
			return true
		}
	}
	return false
}

// TimeUnit 时间单位结构体
type TimeUnit struct {
	Hour   int
	Minute int
	Second int
}
