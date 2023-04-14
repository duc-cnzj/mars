import { AliasToken } from "antd/es/theme/internal";

interface ThemeConfig extends Partial<AliasToken> {
  mainFontColor: string;
  mainColor: string;
  lightColor: string;
  deepColor: string;
  lightMainColor: string;
  mainLinear: string;
  lightLinear: string;
  footerLinear: string;
  titleSecondColor: string;
}

const colors = {
  mainFontColor: "#171717",
  mainColor: "#4f46e5",
  lightColor: "#eef2ff",
  deepColor: "#3730a3",
  lightMainColor: "#a78bfa",
};

const theme: ThemeConfig = {
  colorPrimary: "#4f46e5",
  mainColor: colors.mainColor,
  lightColor: colors.lightColor,
  deepColor: colors.deepColor,
  lightMainColor: colors.lightMainColor,
  mainFontColor: colors.mainFontColor,
  mainLinear: `linear-gradient(to right, ${colors.mainColor}, #34495e)`,
  lightLinear: `linear-gradient(to right, ${colors.lightColor}, #ffffff)`,
  footerLinear: `linear-gradient(to right, ${colors.deepColor}, #34495e)`,
  titleSecondColor: colors.mainColor,
};

export default theme;
