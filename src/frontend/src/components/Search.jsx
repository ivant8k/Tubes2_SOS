'use client';

import React, { useState, useCallback } from 'react';
import dynamic from 'next/dynamic';

// Dynamically import SearchVisualization with no SSR
const SearchVisualization = dynamic(() => import('./SearchVisualization'), {
  ssr: false,
  loading: () => (
    <div className="w-full h-[600px] relative glass rounded-2xl shadow-xl overflow-hidden flex items-center justify-center">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
    </div>
  ),
});

const Search = () => {
  const [searchElement, setSearchElement] = useState('');
  const [searchMode, setSearchMode] = useState('bfs');
  const [isSearching, setIsSearching] = useState(false);
  const [error, setError] = useState(null);
  const [searchResult, setSearchResult] = useState(null);

  const handleSearch = useCallback(async (e) => {
    e.preventDefault();
    if (!searchElement.trim()) return;

    setIsSearching(true);
    setError(null);
    setSearchResult(null);

    try {
      const response = await fetch(
        `http://localhost:5000/search?element=${encodeURIComponent(searchElement)}&mode=${encodeURIComponent(searchMode)}`
      );

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      setSearchResult(data);
    } catch (err) {
      console.error('Search error:', err);
      setError(err.message || 'Error performing search');
    } finally {
      setIsSearching(false);
    }
  }, [searchElement, searchMode]);

  return (
    <div className="space-y-8">
      {/* Search Form */}
      <div className="glass rounded-2xl p-8 shadow-xl">
        <form onSubmit={handleSearch} className="space-y-6">
          <div>
            <label htmlFor="element" className="block text-sm font-medium text-gray-300 mb-2">
              Element to Search
            </label>
            <input
              type="text"
              id="element"
              value={searchElement}
              onChange={(e) => setSearchElement(e.target.value.toLowerCase())}
              placeholder="Enter element name..."
              className="w-full px-4 py-2 bg-gray-800 border border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-white placeholder-gray-400"
              disabled={isSearching}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Search Algorithm
            </label>
            <select
              value={searchMode}
              onChange={(e) => setSearchMode(e.target.value)}
              className="w-full px-4 py-2 bg-gray-800 border border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-white"
              disabled={isSearching}
            >
              <option value="bfs">BFS (Breadth-First Search)</option>
              <option value="dfs">DFS (Depth-First Search)</option>
            </select>
          </div>

          <button
            type="submit"
            disabled={isSearching || !searchElement.trim()}
            className={`w-full py-2 px-4 rounded-lg font-medium transition-colors ${
              isSearching || !searchElement.trim()
                ? 'bg-gray-700 text-gray-400 cursor-not-allowed'
                : 'bg-blue-600 hover:bg-blue-700 text-white'
            }`}
          >
            {isSearching ? 'Searching...' : 'Search'}
          </button>
        </form>
      </div>

      {/* Error Display */}
      {error && (
        <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400">
          {error}
        </div>
      )}

      {/* Search Results */}
      {searchResult && (
        <div className="glass rounded-2xl p-8 shadow-xl">
          <h2 className="text-2xl font-bold mb-4 text-white">Search Results</h2>
          <div className="space-y-4">
            <p className="text-gray-300">
              Element {searchResult.found ? 'found' : 'not found'} after visiting {searchResult.steps} nodes
            </p>
            {searchResult.found && searchResult.path.length > 0 && (
              <div className="space-y-2">
                <h3 className="text-lg font-semibold text-white">Path:</h3>
                <ol className="list-decimal list-inside space-y-1 text-gray-300">
                  {searchResult.path.map((step, index) => (
                    <li key={index}>
                      {step.ingredients[0]} + {step.ingredients[1]} = {step.result}
                    </li>
                  ))}
                </ol>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Visualization */}
      {searchElement && (
        <div className="glass rounded-2xl p-8 shadow-xl">
          <h2 className="text-2xl font-bold mb-4 text-white">Visualization</h2>
          <SearchVisualization 
            element={searchElement} 
            mode={searchMode} 
            solutionPath={searchResult?.path || []}
          />
        </div>
      )}
    </div>
  );
};

export default Search; 