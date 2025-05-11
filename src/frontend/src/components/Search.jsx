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
  const [selectedRecipeIndex, setSelectedRecipeIndex] = useState(0);

  const handleSearch = useCallback(async (e) => {
    e.preventDefault();
    if (!searchElement.trim()) return;

    setIsSearching(true);
    setError(null);
    setSearchResult(null);
    setSelectedRecipeIndex(0);

    try {
      const response = await fetch(
        `http://localhost:5000/search?element=${encodeURIComponent(searchElement)}&mode=${encodeURIComponent(searchMode)}`
      );

      if (!response.ok) {
        if (response.status === 404) {
          throw new Error(`Element "${searchElement}" not found`);
        }
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

  const handleRecipeChange = (index) => {
    setSelectedRecipeIndex(index);
  };

  return (
    <div className="container mx-auto px-4 py-8 space-y-8">
      {/* Search Form */}
      <form onSubmit={handleSearch} className="glass rounded-2xl p-4 sm:p-8 shadow-xl">
        <div className="flex flex-col md:flex-row gap-4">
          <input
            type="text"
            value={searchElement}
            onChange={(e) => setSearchElement(e.target.value)}
            placeholder="Enter element to search..."
            className="flex-1 px-4 py-2 rounded-lg bg-white/10 border border-white/20 text-white placeholder-gray-400 focus:outline-none focus:border-blue-500 text-sm sm:text-base"
          />
          <div className="relative">
            <select
              value={searchMode}
              onChange={(e) => setSearchMode(e.target.value)}
              className="w-full sm:w-auto appearance-none px-4 py-2 rounded-lg bg-white/10 border border-white/20 text-white focus:outline-none focus:border-blue-500 pr-10 cursor-pointer hover:bg-white/15 transition-colors text-sm sm:text-base"
            >
              <option value="bfs" className="bg-gray-800 text-white">BFS</option>
              <option value="dfs" className="bg-gray-800 text-white">DFS</option>
              <option value="multi" className="bg-gray-800 text-white">Multi-Recipe</option>
            </select>
            <div className="absolute inset-y-0 right-0 flex items-center px-2 pointer-events-none">
              <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </div>
          </div>
          <button
            type="submit"
            disabled={isSearching}
            className="w-full sm:w-auto px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors text-sm sm:text-base"
          >
            {isSearching ? 'Searching...' : 'Search'}
          </button>
        </div>
      </form>

      {/* Error Display */}
      {error && (
        <div className="p-3 sm:p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 text-sm sm:text-base">
          {error}
        </div>
      )}

      {/* Search Results */}
      {searchResult && (
        <div className="glass rounded-2xl p-4 sm:p-8 shadow-xl">
          <h2 className="text-xl sm:text-2xl font-bold mb-4 text-white">Search Results</h2>
          <div className="space-y-4">
            <p className="text-gray-300 text-sm sm:text-base">
              Element {searchResult.found ? 'found' : 'not found'} after visiting {searchResult.steps} nodes
            </p>
            {searchResult.found && searchResult.paths && searchResult.paths.length > 0 && (
              <div className="space-y-4">
                {/* Target Element Info */}
                <div className="bg-white/5 p-3 rounded-lg">
                  <p className="text-white text-sm sm:text-base">
                    Target: <span className="font-semibold">{searchResult.target.element}</span>
                    <span className="ml-2 px-2 py-0.5 bg-blue-500/20 text-blue-300 rounded text-xs">
                      Tier {searchResult.target.tier}
                    </span>
                  </p>
                </div>

                {/* Recipe Selector */}
                {searchResult.paths.length > 1 && (
                  <div className="flex flex-wrap gap-2 mb-6">
                    {searchResult.paths.map((_, index) => (
                      <button
                        key={index}
                        onClick={() => handleRecipeChange(index)}
                        className={`px-3 sm:px-4 py-1.5 sm:py-2 rounded-lg transition-all duration-200 text-sm sm:text-base ${
                          selectedRecipeIndex === index
                            ? 'bg-blue-500 text-white shadow-lg scale-105'
                            : 'bg-white/10 text-gray-300 hover:bg-white/20'
                        }`}
                      >
                        Recipe {index + 1}
                      </button>
                    ))}
                  </div>
                )}
                
                {/* Selected Recipe Path */}
                <div className="space-y-2">
                  <h3 className="text-base sm:text-lg font-semibold text-white">
                    {searchResult.paths.length > 1 ? `Recipe ${selectedRecipeIndex + 1} Path:` : 'Path:'}
                  </h3>
                  <ol className="list-decimal list-inside space-y-1 text-gray-300 text-sm sm:text-base">
                    {searchResult.paths[selectedRecipeIndex].map((step, index) => (
                      <li key={index} className="hover:bg-white/5 p-2 rounded transition-colors">
                        <div className="flex items-center gap-2">
                          <span className="font-medium">{step.ingredients[0]}</span>
                          <span className="px-1.5 py-0.5 bg-gray-700/50 text-gray-300 rounded text-xs">
                            Tier {step.tiers.left}
                          </span>
                          <span>+</span>
                          <span className="font-medium">{step.ingredients[1]}</span>
                          <span className="px-1.5 py-0.5 bg-gray-700/50 text-gray-300 rounded text-xs">
                            Tier {step.tiers.right}
                          </span>
                          <span>=</span>
                          <span className="font-medium">{step.result}</span>
                          <span className="px-1.5 py-0.5 bg-blue-500/20 text-blue-300 rounded text-xs">
                            Tier {step.tiers.result}
                          </span>
                        </div>
                      </li>
                    ))}
                  </ol>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Visualization */}
      {searchElement && searchResult?.paths && (
        <div className="glass rounded-2xl p-4 sm:p-8 shadow-xl">
          <h2 className="text-xl sm:text-2xl font-bold mb-4 text-white">Visualization</h2>
          <SearchVisualization 
            element={searchElement} 
            mode={searchMode} 
            solutionPath={searchResult.paths[selectedRecipeIndex] || []}
          />
        </div>
      )}
    </div>
  );
};

export default Search; 