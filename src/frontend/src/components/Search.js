import React, { useState } from 'react';
import SearchVisualization from './SearchVisualization';

const Search = () => {
  const [element, setElement] = useState('');
  const [mode, setMode] = useState('bfs');
  const [result, setResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [showVisualization, setShowVisualization] = useState(false);

  const handleSearch = async () => {
    if (!element) {
      setError('Please enter an element');
      return;
    }

    setLoading(true);
    setError(null);
    setResult(null);
    setShowVisualization(true);

    try {
      const response = await fetch(
        `http://localhost:5000/search?element=${encodeURIComponent(element)}&mode=${mode}`
      );
      const data = await response.json();
      setResult(data);
    } catch (err) {
      setError('Error searching for element');
      console.error('Search error:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="bg-white rounded-lg shadow-lg p-6 mb-6">
        <h2 className="text-2xl font-bold mb-4">Search for Elements</h2>
        <div className="flex gap-4 mb-4">
          <input
            type="text"
            value={element}
            onChange={(e) => setElement(e.target.value)}
            placeholder="Enter element name"
            className="flex-1 px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <select
            value={mode}
            onChange={(e) => setMode(e.target.value)}
            className="px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="bfs">BFS</option>
            <option value="dfs">DFS</option>
          </select>
          <button
            onClick={handleSearch}
            disabled={loading}
            className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
          >
            {loading ? 'Searching...' : 'Search'}
          </button>
        </div>
        {error && <div className="text-red-500 mb-4">{error}</div>}
      </div>

      {showVisualization && (
        <div className="bg-white rounded-lg shadow-lg p-6 mb-6">
          <h2 className="text-2xl font-bold mb-4">Search Visualization</h2>
          <SearchVisualization element={element} mode={mode} />
        </div>
      )}

      {result && (
        <div className="bg-white rounded-lg shadow-lg p-6">
          <h2 className="text-2xl font-bold mb-4">Search Results</h2>
          {result.found ? (
            <div>
              <p className="mb-4">
                Found {result.path.length} steps to create {element}:
              </p>
              <div className="space-y-2">
                {result.path.map((step, index) => (
                  <div
                    key={index}
                    className="p-3 bg-gray-50 rounded-lg flex items-center gap-2"
                  >
                    <span className="font-medium">{step.ingredients[0]}</span>
                    <span>+</span>
                    <span className="font-medium">{step.ingredients[1]}</span>
                    <span>=</span>
                    <span className="font-medium text-blue-500">
                      {step.result}
                    </span>
                  </div>
                ))}
              </div>
              <p className="mt-4 text-gray-600">
                Total nodes visited: {result.steps}
              </p>
            </div>
          ) : (
            <p className="text-red-500">
              Could not find a way to create {element}
            </p>
          )}
        </div>
      )}
    </div>
  );
};

export default Search; 