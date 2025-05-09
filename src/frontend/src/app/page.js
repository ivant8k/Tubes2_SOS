'use client';

import { useState } from 'react';
import Tree from 'react-d3-tree';
import axios from 'axios';

// Create axios instance with default config
const api = axios.create({
  baseURL: 'http://localhost:5000',
  timeout: 10000, // 10 seconds
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  },
  withCredentials: false
});

export default function Home() {
  const [searchMode, setSearchMode] = useState('bfs'); // 'bfs' or 'dfs'
  const [searchTerm, setSearchTerm] = useState('');
  const [treeData, setTreeData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [stats, setStats] = useState(null);

  const handleSearch = async () => {
    if (!searchTerm) return;
    
    setLoading(true);
    setError(null);
    setStats(null);
    
    try {
      console.log('Sending request to:', `${api.defaults.baseURL}/search`);
      const response = await api.get('/search', {
        params: {
          element: searchTerm,
          mode: searchMode
        }
      });
      
      console.log('Received response:', response.data);
      
      if (!response.data.found) {
        setError('Element not found. Please try another element.');
        setTreeData(null);
        return;
      }

      // Transform the response data into the format expected by react-d3-tree
      const transformedData = transformDataForTree(response.data.path);
      setTreeData(transformedData);
      setStats({
        steps: response.data.steps,
        pathLength: response.data.path.length
      });
    } catch (err) {
      console.error('Search error:', err);
      if (err.code === 'ECONNABORTED') {
        setError('Request timed out. Please try again.');
      } else if (!err.response) {
        setError('Cannot connect to server. Please make sure the backend is running on port 5000.');
      } else {
        setError(`Error: ${err.message}`);
      }
      setTreeData(null);
    } finally {
      setLoading(false);
    }
  };

  const transformDataForTree = (path) => {
    if (!path || path.length === 0 || !path[0] || !path[0].ingredients) {
      return null;
    }

    // Create a tree structure from the path
    const root = {
      name: path[0].ingredients[0],
      children: []
    };

    let currentNode = root;
    
    for (const step of path) {
      const newNode = {
        name: step.result,
        children: []
      };
      
      // Add ingredients as children
      newNode.children.push({
        name: step.ingredients[0],
        children: []
      });
      newNode.children.push({
        name: step.ingredients[1],
        children: []
      });
      
      currentNode.children.push(newNode);
      currentNode = newNode;
    }

    return root;
  };

  return (
    <main className="min-h-screen bg-gray-100 p-8">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-4xl font-bold text-center mb-8 text-gray-800">
          Little Alchemy 2 Path Finder
        </h1>
        
        <div className="bg-white rounded-lg shadow-lg p-6 mb-8">
          <div className="flex flex-col space-y-4">
            <div className="flex items-center space-x-4">
              <input
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="Enter element to search..."
                className="flex-1 p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
              
              <div className="flex space-x-2">
                <button
                  onClick={() => setSearchMode('bfs')}
                  className={`px-4 py-2 rounded-lg ${
                    searchMode === 'bfs'
                      ? 'bg-blue-500 text-white'
                      : 'bg-gray-200 text-gray-700'
                  }`}
                >
                  BFS
                </button>
                <button
                  onClick={() => setSearchMode('dfs')}
                  className={`px-4 py-2 rounded-lg ${
                    searchMode === 'dfs'
                      ? 'bg-blue-500 text-white'
                      : 'bg-gray-200 text-gray-700'
                  }`}
                >
                  DFS
                </button>
              </div>
              
              <button
                onClick={handleSearch}
                disabled={loading}
                className="px-6 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 disabled:bg-gray-400"
              >
                {loading ? 'Searching...' : 'Search'}
              </button>
            </div>
            
            {error && (
              <div className="text-red-500 text-sm">{error}</div>
            )}

            {stats && (
              <div className="text-sm text-gray-600">
                <p>Nodes visited: {stats.steps}</p>
                <p>Path length: {stats.pathLength} steps</p>
              </div>
            )}
          </div>
        </div>

        {treeData && (
          <div className="bg-white rounded-lg shadow-lg p-6">
            <div className="h-[600px] w-full">
              <Tree
                data={treeData}
                orientation="vertical"
                pathFunc="step"
                separation={{ siblings: 2, nonSiblings: 2.5 }}
                renderCustomNodeElement={({ nodeDatum }) => (
                  <g>
                    <circle r={15} fill="#4F46E5" />
                    <text
                      dy=".31em"
                      x={20}
                      textAnchor="start"
                      style={{ fill: '#1F2937' }}
                    >
                      {nodeDatum.name}
                    </text>
                  </g>
                )}
              />
            </div>
          </div>
        )}
    </div>
    </main>
  );
}
