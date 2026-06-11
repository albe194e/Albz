import 'dart:math' as math;

class AppBreakpoints {
  static const double mobile = 760;
  static const double navRail = 900;
  static const double narrow = 520;

  static bool isMobile(double width) => width < mobile;

  static bool showNavRail(double width) => width >= navRail;

  static bool isNarrow(double width) => width < narrow;

  static double mobileDrawerWidth(double width) {
    return math.min(360, width * 0.88);
  }
}
