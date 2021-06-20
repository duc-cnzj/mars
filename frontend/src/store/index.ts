import { createStore, applyMiddleware } from "redux";
import rootReducer from "./reducers";
import { composeWithDevTools } from 'redux-devtools-extension';
import thunk from 'redux-thunk';

const composeEnhancers = composeWithDevTools({});

const enhancers = process.env.NODE_ENV === "production" ? applyMiddleware(thunk) : composeEnhancers(
  applyMiddleware(thunk),
);

export default createStore(rootReducer, enhancers);