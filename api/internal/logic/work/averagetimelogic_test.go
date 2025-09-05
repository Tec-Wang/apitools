package work

import (
	"context"
	"testing"
	"time"

	"apitools/api/internal/consts"
	"apitools/api/internal/svc"
	"apitools/api/internal/types"
)

func TestAverageTimeLogic_AverageTimeHourMinuteSecond(t *testing.T) {
	// 创建测试用的服务上下文
	svcCtx := &svc.ServiceContext{}
	ctx := context.Background()
	logic := NewAverageTimeLogic(ctx, svcCtx)

	tests := []struct {
		name           string
		req            *types.AverageTimeReq
		expectedResult string // 期望的结果时间（格式：HH:MM:SS）
		expectError    bool
	}{
		{
			name: "测试基本平均时间计算 - 时分秒精度",
			req: &types.AverageTimeReq{
				TimestampList: []int64{
					createTimestamp(12, 0, 0), // 12:00:00
					createTimestamp(12, 2, 0), // 12:02:00
				},
				CalculateType: consts.HourMinuteSecond,
			},
			expectedResult: "12:01:00",
			expectError:    false,
		},
		{
			name: "测试三个时间点的平均 - 时分秒精度",
			req: &types.AverageTimeReq{
				TimestampList: []int64{
					createTimestamp(9, 0, 0),  // 09:00:00
					createTimestamp(9, 15, 0), // 09:15:00
					createTimestamp(9, 30, 0), // 09:30:00
				},
				CalculateType: consts.HourMinuteSecond,
			},
			expectedResult: "09:15:00",
			expectError:    false,
		},
		{
			name: "测试带秒精度的时间",
			req: &types.AverageTimeReq{
				TimestampList: []int64{
					createTimestamp(14, 30, 15), // 14:30:15
					createTimestamp(14, 30, 45), // 14:30:45
				},
				CalculateType: consts.HourMinuteSecond,
			},
			expectedResult: "14:30:30",
			expectError:    false,
		},
		{
			name: "测试空时间戳列表",
			req: &types.AverageTimeReq{
				TimestampList: []int64{},
				CalculateType: consts.HourMinuteSecond,
			},
			expectedResult: "",
			expectError:    true,
		},
		{
			name: "测试单个时间点",
			req: &types.AverageTimeReq{
				TimestampList: []int64{
					createTimestamp(16, 45, 30), // 16:45:30
				},
				CalculateType: consts.HourMinuteSecond,
			},
			expectedResult: "16:45:30",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := logic.AverageTimeHourMinuteSecond(tt.req)

			if tt.expectError {
				if err == nil {
					t.Errorf("期望出现错误，但没有错误")
				}
				return
			}

			if err != nil {
				t.Errorf("不期望出现错误，但出现了错误: %v", err)
				return
			}

			if resp == nil {
				t.Errorf("响应不能为nil")
				return
			}

			// 将结果时间戳转换为时间格式进行比较
			resultTime := time.Unix(resp.AverageTime, 0)
			resultStr := resultTime.Format("15:04:05")

			if resultStr != tt.expectedResult {
				t.Errorf("期望结果 %s，实际结果 %s", tt.expectedResult, resultStr)
			}
		})
	}
}

// createTimestamp 创建指定时分秒的时间戳（使用今天的日期）
func createTimestamp(hour, minute, second int) int64 {
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, now.Location())
	return t.Unix()
}

func TestTimeUnit_hasSecondPrecision(t *testing.T) {
	logic := &AverageTimeLogic{}

	tests := []struct {
		name     string
		units    []TimeUnit
		expected bool
	}{
		{
			name: "有秒精度",
			units: []TimeUnit{
				{Hour: 12, Minute: 0, Second: 0},
				{Hour: 12, Minute: 2, Second: 30},
			},
			expected: true,
		},
		{
			name: "无秒精度",
			units: []TimeUnit{
				{Hour: 12, Minute: 0, Second: 0},
				{Hour: 12, Minute: 2, Second: 0},
			},
			expected: false,
		},
		{
			name: "部分有秒精度",
			units: []TimeUnit{
				{Hour: 12, Minute: 0, Second: 0},
				{Hour: 12, Minute: 2, Second: 15},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logic.hasSecondPrecision(tt.units)
			if result != tt.expected {
				t.Errorf("期望 %v，实际 %v", tt.expected, result)
			}
		})
	}
}
