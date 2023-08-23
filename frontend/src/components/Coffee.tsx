import React, { memo } from "react";
import Alipay from "../assets/alipay.jpg";
import WechatPay from "../assets/wechatpay.jpg";
import styled from "@emotion/styled";
import theme from "../styles/theme";

const Coffee: React.FC = () => {
  return (
    <Wrap>
      <Front className="front">
        <Image src={Alipay} alt="alipay" />
      </Front>
      <Back className="back">
        <Image src={WechatPay} alt="wechatpay" />
      </Back>
    </Wrap>
  );
};

export default memo(Coffee);

const Item = styled.div`
  position: absolute;
  backface-visibility: hidden;
  transition: 0.3s;
`;

const Front = styled(Item)`
  transform: rotateX(0deg);
`;

const Back = styled(Item)`
  transform: rotateX(-180deg);
`;

const Wrap = styled.div`
  perspective: 700px;
  width: 200px;
  height: 273px;
  margin: 0 auto;

  &:hover .front {
    outline: 1px solid blue;
    transform: rotateX(180deg);
  }

  &:hover .back {
    transform: rotateX(0deg);
  }
`;

const Image = styled.img`
  width: 200px;
  height: 273px;
  border-radius: 5px;
  box-shadow: 1px 1px 10px ${theme.lightMainColor};
`;
