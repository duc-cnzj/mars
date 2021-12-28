import { Component, createRef } from 'react';
import { bool, func, number } from 'prop-types';
import FilterLinesIcon from './FilterLinesIcon';
import { SEARCH_MIN_KEYWORDS } from '../../utils';
import styles from './index.module.css';
import hotkeys from 'hotkeys-js';

export default class SearchBar extends Component {
  static propTypes = {
    /**
     * Executes a function when the user starts typing.
     */
    onSearch: func,
    /**
     * Executes a function when the search input has been cleared.
     */
    onClearSearch: func,
    /**
     * Executes a function when the option `Filter Lines With Matches`
     * is enable.
     */
    onFilterLinesWithMatches: func,
    /**
     * Number of search results. Should come from the component
     * executing the search algorithm.
     */
    resultsCount: number,
    /**
     * If true, then only lines that match the search term will be displayed.
     */
    filterActive: bool,
    /**
     * If true, the input field and filter button will be disabled.
     */
    disabled: bool,
    /**
   * If true, capture system hotkeys for searching the page (Cmd-F, Ctrl-F,
   * etc.)
   */
    captureHotkeys: bool,
  };

  static defaultProps = {
    onSearch: () => {},
    onClearSearch: () => {},
    onFilterLinesWithMatches: () => {},
    resultsCount: 0,
    filterActive: false,
    disabled: false,
    captureHotkeys: false,
  };

  handleSearchHotkey = e => {
    if (!this.inputRef.current) {
      return;
    }

    e.preventDefault();
    this.inputRef.current.focus();
  };

  componentDidMount() {
    if (this.props.captureHotkeys) {
      hotkeys('ctrl+f,cmd+f', this.handleSearchHotkey);
    }
  }

  constructor(props) {
    super(props);
    this.inputRef = createRef();
  }

  state = {
    keywords: '',
  };

  handleFilterToggle = () => {
    this.props.onFilterLinesWithMatches(!this.props.filterActive);
  };

  handleSearchChange = e => {
    const { value: keywords } = e.target;

    this.setState({ keywords }, () => this.search());
  };

  handleSearchKeyPress = e => {
    if (e.key === 'Enter') {
      this.handleFilterToggle();
    }
  };

  search = () => {
    const { keywords } = this.state;
    const { onSearch, onClearSearch } = this.props;

    if (keywords && keywords.length > SEARCH_MIN_KEYWORDS) {
      onSearch(keywords);
    } else {
      onClearSearch();
    }
  };

  render() {
    const { resultsCount, filterActive, disabled } = this.props;
    const matchesLabel = `match${resultsCount === 1 ? '' : 'es'}`;
    const filterIcon = filterActive ? styles.active : styles.inactive;

    return (
      <div className={`react-lazylog-searchbar ${styles.searchBar}`}>
        <input
          autoComplete="off"
          type="text"
          name="search"
          placeholder="Search"
          className={`react-lazylog-searchbar-input ${styles.searchInput}`}
          onChange={this.handleSearchChange}
          onKeyPress={this.handleSearchKeyPress}
          value={this.state.keywords}
          disabled={disabled}
          ref={this.inputRef}
        />
        <button
          disabled={disabled}
          className={`react-lazylog-searchbar-filter ${
            filterActive ? 'active' : 'inactive'
          } ${styles.button} ${filterIcon}`}
          onClick={this.handleFilterToggle}>
          <FilterLinesIcon />
        </button>
        <span
          className={`react-lazylog-searchbar-matches ${
            resultsCount ? 'active' : 'inactive'
          } ${resultsCount ? styles.active : styles.inactive}`}>
          {resultsCount} {matchesLabel}
        </span>
      </div>
    );
  }
}
