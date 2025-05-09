'use client';

import React, { Suspense } from 'react';
import dynamic from 'next/dynamic';

// Dynamically import components with no SSR
const Search = dynamic(() => import('../components/Search'), {
  ssr: false,
  loading: () => (
    <div className="w-full h-[600px] relative glass rounded-2xl shadow-xl overflow-hidden flex items-center justify-center">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
    </div>
  ),
});

export default function Home() {
  return (
    <main className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
      <div className="container mx-auto px-4 py-8">
        {/* Hero Section */}
        <div className="text-center mb-12">
          <h1 className="text-4xl md:text-6xl font-bold mb-4 bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-purple-500">
            Element Search Visualization
          </h1>
          <p className="text-xl text-gray-300 max-w-2xl mx-auto">
            Visualize how BFS and DFS algorithms search for element combinations in real-time
          </p>
        </div>

        {/* Search Section */}
        <div className="max-w-4xl mx-auto">
          <Suspense fallback={
            <div className="w-full h-[600px] relative glass rounded-2xl shadow-xl overflow-hidden flex items-center justify-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
            </div>
          }>
            <Search />
          </Suspense>
        </div>

        {/* Footer */}
        <footer className="mt-12 text-center text-gray-400">
          <p>Kelompok SOS</p>
        </footer>
      </div>
    </main>
  );
} 