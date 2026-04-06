/**
 * Convenience constants for converting between different time units.
 */
export class TimeInSeconds {
  static MS_PER_SECOND = 1000;
  static SECOND = 1;
  static MINUTE = TimeInSeconds.SECOND * 60;
  static HOUR = TimeInSeconds.MINUTE * 60;
  static DAY = TimeInSeconds.HOUR * 24;
}
