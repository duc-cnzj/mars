import { createFromIconfontCN } from "@ant-design/icons";

const MyIcon = createFromIconfontCN({
  scriptUrl: "at.alicdn.com/t/c/font_4221170_xal0mzzssng.css",
});

type Props = {
  name: string;
};

const Icon: React.FC<Props> = (props) => {
  return <MyIcon type={`icon-${props.name}`} />;
};
//at.alicdn.com/t/c/font_4221170_xal0mzzssng.css
export default Icon;
