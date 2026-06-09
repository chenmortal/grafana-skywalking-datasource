import { getTemplateSrv, locationService } from '@grafana/runtime';
import { useEffect, useState } from 'react';

/**
 * 时间范围接口
 */
export interface TimeRangeValues {
  from: number;
  to: number;
}

/**
 * 获取当前时间范围
 * @returns 时间范围对象，包含 from 和 to（毫秒时间戳）
 */
export function getTimeRangeValues(): TimeRangeValues {
  const templateSrv = getTemplateSrv();
  return {
    from: Number(templateSrv.replace('${__from}')),
    to: Number(templateSrv.replace('${__to}')),
  };
}

/**
 * 获取当前时间范围（秒）
 * @returns 时间范围对象，包含 from 和 to（秒时间戳）
 */
export function getTimeRangeInSeconds(): { from: number; to: number } {
  const { from, to } = getTimeRangeValues();
  return {
    from: Math.floor(from / 1000),
    to: Math.floor(to / 1000),
  };
}

/**
 * 获取时间范围的持续时间（毫秒）
 * @returns 持续时间（毫秒）
 */
export function getTimeRangeDuration(): number {
  const { from, to } = getTimeRangeValues();
  return to - from;
}

export function useTimeRangeFromUrl() {
  const [timeRange, setTimeRange] = useState<TimeRangeValues>(() => {
    const time_range = getTimeRangeValues();
    return time_range;
  });

  useEffect(() => {
    const subscription = locationService.getLocationObservable().subscribe((_) => {
      setTimeRange(getTimeRangeValues());
    });
    return () => subscription.unsubscribe();
  }, []);

  return timeRange;
}

/**
 * 解码服务ID
 * @param id 编码后的服务ID
 * @returns 解码后的服务ID，解码失败则返回原值
 */
export function descodeServiceID(id: string): string {
  try {
    const encodeName=id.split(".1")[0]
    return atob(encodeName);
  } catch {
    return id;
  }
}

