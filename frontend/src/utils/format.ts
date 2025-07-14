import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import 'dayjs/locale/zh-cn';

dayjs.extend(relativeTime);
dayjs.locale('zh-cn');

// 格式化日期
export const formatDate = (date: string | Date, format: string = 'YYYY-MM-DD HH:mm:ss') => {
  return dayjs(date).format(format);
};

// 相对时间
export const formatRelativeTime = (date: string | Date) => {
  return dayjs(date).fromNow();
};

// 格式化数字
export const formatNumber = (num: number, decimals: number = 2): string => {
  return num.toFixed(decimals);
};

// 格式化金额
export const formatCurrency = (amount: number, decimals: number = 2): string => {
  return new Intl.NumberFormat('zh-CN', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(amount);
};

// 格式化百分比
export const formatPercent = (value: number, decimals: number = 2): string => {
  return `${(value * 100).toFixed(decimals)}%`;
};

// 格式化文件大小
export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes';
  
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

// 缩略显示地址
export const formatAddress = (address: string, start: number = 6, end: number = 4): string => {
  if (!address) return '';
  if (address.length <= start + end) return address;
  return `${address.slice(0, start)}...${address.slice(-end)}`;
};

// 格式化交易对
export const formatSymbol = (symbol: string): string => {
  return symbol.replace(/USDT$/, '/USDT').replace(/BUSD$/, '/BUSD');
};