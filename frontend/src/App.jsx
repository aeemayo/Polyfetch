import React, { useState, useEffect, useCallback } from 'react';
import axios from 'axios';
import './App.css';

const API_BASE = 'http://localhost:8080/api';

// Icons as simple SVG components
const SearchIcon = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <circle cx="11" cy="11" r="8"/>
    <path d="m21 21-4.35-4.35"/>
  </svg>
);

const ChartIcon = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M18 20V10M12 20V4M6 20v-6"/>
  </svg>
);

const UsersIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/>
    <circle cx="9" cy="7" r="4"/>
    <path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"/>
  </svg>
);

const VolumeIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>
  </svg>
);

const RefreshIcon = ({ spinning }) => (
  <svg 
    width="18" 
    height="18" 
    viewBox="0 0 24 24" 
    fill="none" 
    stroke="currentColor" 
    strokeWidth="2"
    style={{ animation: spinning ? 'spin 1s linear infinite' : 'none' }}
  >
    <path d="M21 12a9 9 0 1 1-9-9c2.52 0 4.93 1 6.74 2.74L21 8"/>
    <path d="M21 3v5h-5"/>
  </svg>
);

const TrendingIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/>
    <polyline points="17 6 23 6 23 12"/>
  </svg>
);

function App() {
  const [markets, setMarkets] = useState([]);
  const [selectedMarket, setSelectedMarket] = useState(null);
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [statsLoading, setStatsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');

  // Fetch markets on mount
  const fetchMarkets = useCallback(async (search = '') => {
    setLoading(true);
    setError(null);
    try {
      const endpoint = search 
        ? `${API_BASE}/markets/search?q=${encodeURIComponent(search)}&limit=30`
        : `${API_BASE}/markets?limit=30`;
      
      const response = await axios.get(endpoint);
      if (response.data.success) {
        setMarkets(response.data.data || []);
      } else {
        setError(response.data.error || 'Failed to fetch markets');
      }
    } catch (err) {
      console.error('Error fetching markets:', err);
      setError(err.message || 'Failed to connect to server');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchMarkets();
  }, [fetchMarkets]);

  // Debounced search
  useEffect(() => {
    const timer = setTimeout(() => {
      fetchMarkets(searchQuery);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery, fetchMarkets]);

  // Fetch stats when a market is selected
  const handleSelectMarket = async (market) => {
    setSelectedMarket(market);
    setStatsLoading(true);
    setStats(null);
    
    try {
      const response = await axios.get(`${API_BASE}/market/${market.id}/stats`);
      if (response.data.success) {
        setStats(response.data.data);
      }
    } catch (err) {
      console.error('Error fetching stats:', err);
    } finally {
      setStatsLoading(false);
    }
  };

  // Format volume for display
  const formatVolume = (volume) => {
    if (!volume) return '$0';
    const num = parseFloat(volume);
    if (num >= 1000000) return `$${(num / 1000000).toFixed(1)}M`;
    if (num >= 1000) return `$${(num / 1000).toFixed(1)}K`;
    return `$${num.toFixed(0)}`;
  };

  // Format percentage
  const formatPercentage = (pct) => {
    return `${pct.toFixed(1)}%`;
  };

  // Get percentage class for styling
  const getPercentageClass = (pct) => {
    if (pct >= 50) return 'high';
    if (pct >= 25) return 'medium';
    return 'low';
  };

  // Get leading outcome from market
  const getLeadingOutcome = (market) => {
    if (!market.outcomes || !market.outcomePrices) return null;
    const outcomes = market.outcomes;
    const prices = market.outcomePrices;
    
    let maxIdx = 0;
    let maxPrice = 0;
    
    prices.forEach((price, idx) => {
      const p = parseFloat(price) || 0;
      if (p > maxPrice) {
        maxPrice = p;
        maxIdx = idx;
      }
    });
    
    return {
      name: outcomes[maxIdx],
      price: maxPrice
    };
  };

  return (
    <div className="app-container">
      {/* Header */}
      <header className="app-header">
        <div className="app-logo">
          <div className="logo-icon">📊</div>
          <h1 className="app-title">Polyfetch</h1>
        </div>
        <p className="app-subtitle">
          Real-time Polymarket analysis • Identify popular outcomes & betting trends
        </p>
      </header>

      {/* Search bar */}
      <div className="search-container">
        <div className="search-input-wrapper">
          <span className="search-icon"><SearchIcon /></span>
          <input
            type="text"
            className="search-input"
            placeholder="Search markets (e.g., 'Bitcoin', 'Election', 'AI')"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
      </div>

      {/* Main layout */}
      <div className="main-layout">
        {/* Markets list */}
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">
              <ChartIcon /> Active Markets
            </h2>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
              {!loading && <span className="card-badge">{markets.length}</span>}
              <button 
                className="refresh-btn" 
                onClick={() => fetchMarkets(searchQuery)}
                disabled={loading}
                title="Refresh markets"
              >
                <RefreshIcon spinning={loading} />
              </button>
            </div>
          </div>
          <div className="card-content">
            {loading ? (
              <div className="loading-container">
                <div className="loading-spinner"></div>
                <p className="loading-text">Loading markets...</p>
              </div>
            ) : error ? (
              <div className="error-container">
                <div className="error-icon">⚠️</div>
                <p>{error}</p>
                <button 
                  className="refresh-btn" 
                  onClick={() => fetchMarkets()}
                  style={{ margin: '1rem auto' }}
                >
                  Try Again
                </button>
              </div>
            ) : markets.length === 0 ? (
              <div className="empty-state">
                <div className="empty-icon">🔍</div>
                <p>No markets found</p>
              </div>
            ) : (
              <div className="market-list">
                {markets.map((market, idx) => {
                  const leading = getLeadingOutcome(market);
                  return (
                    <div
                      key={market.id || idx}
                      className={`market-item animate-fade-in ${selectedMarket?.id === market.id ? 'active' : ''}`}
                      onClick={() => handleSelectMarket(market)}
                      style={{ animationDelay: `${idx * 50}ms` }}
                    >
                      <p className="market-question">{market.question}</p>
                      <div className="market-meta">
                        <span className="market-meta-item">
                          <VolumeIcon /> {formatVolume(market.volume)}
                        </span>
                        {market.outcomes && (
                          <span className="market-meta-item">
                            {market.outcomes.length} outcomes
                          </span>
                        )}
                      </div>
                      {leading && market.outcomes && (
                        <div className="market-outcomes-preview">
                          {market.outcomes.slice(0, 3).map((outcome, i) => {
                            const price = market.outcomePrices?.[i] || '0';
                            const isLeading = outcome === leading.name;
                            return (
                              <span 
                                key={i} 
                                className={`outcome-chip ${isLeading ? 'leading' : ''}`}
                              >
                                {outcome}
                                <span className="price">
                                  {(parseFloat(price) * 100).toFixed(0)}¢
                                </span>
                              </span>
                            );
                          })}
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>

        {/* Stats panel */}
        <div className="card stats-panel">
          {!selectedMarket ? (
            <div className="stats-empty">
              <div className="stats-empty-icon">📈</div>
              <p>Select a market to view betting statistics</p>
            </div>
          ) : statsLoading ? (
            <div className="loading-container">
              <div className="loading-spinner"></div>
              <p className="loading-text">Analyzing market data...</p>
            </div>
          ) : stats ? (
            <>
              <div className="stats-header">
                <h3 className="stats-question">{stats.question || selectedMarket.question}</h3>
                <div className="stats-summary">
                  <div className="stat-box" style={{ flex: 1 }}>
                    <div className="stat-value" style={{ color: 'var(--success)' }}>
                      <TrendingIcon style={{ marginRight: '0.25rem' }} />
                      {formatPercentage(stats.popularPct)}
                    </div>
                    <div className="stat-label">Consensus Probability</div>
                  </div>
                </div>
              </div>
              
              <div className="outcomes-list">
                <h4 style={{ 
                  fontSize: '0.9rem', 
                  color: 'var(--text-secondary)', 
                  marginBottom: '1.25rem',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em'
                }}>
                  Outcome Probabilities
                </h4>
                
                {stats.outcomeStats?.map((outcome, idx) => {
                  const pctClass = getPercentageClass(outcome.percentage);
                  const isPopular = outcome.outcome === stats.popularOutcome;
                  
                  return (
                    <div 
                      key={idx} 
                      className="outcome-item animate-fade-in"
                      style={{ animationDelay: `${idx * 100}ms` }}
                    >
                      <div className="outcome-header">
                        <span className="outcome-name">
                          {outcome.outcome}
                          {isPopular && <span className="popular-badge">Consensus Pick</span>}
                        </span>
                        <span className={`outcome-percentage ${pctClass}`}>
                          {formatPercentage(outcome.percentage)}
                        </span>
                      </div>
                      <div className="outcome-bar-container">
                        <div 
                          className={`outcome-bar ${pctClass}`}
                          style={{ width: `${Math.max(outcome.percentage, 2)}%` }}
                        />
                      </div>
                      <div className="outcome-details">
                        <span>
                           Implied Probability: {formatPercentage(outcome.percentage)}
                        </span>
                        {outcome.price && (
                          <span className="price-tag">
                            Price: {(parseFloat(outcome.price) * 100).toFixed(1)}¢
                          </span>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            </>
          ) : (
            <div className="error-container">
              <p>Failed to load statistics</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default App;
