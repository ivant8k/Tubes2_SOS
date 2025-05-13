"use client";

import React, { Suspense } from 'react';
import dynamic from 'next/dynamic';
import { motion } from 'framer-motion';

// 3D Element SVG (same as landing page)
const ThreeDElement = () => (
  <motion.svg
    width="180"
    height="180"
    viewBox="0 0 220 220"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    initial={{ rotate: -10, scale: 0.8 }}
    animate={{ rotate: 10, scale: 1 }}
    transition={{ yoyo: Infinity, duration: 2, ease: 'easeInOut' }}
    className="drop-shadow-2xl mb-2"
  >
    <defs>
      <radialGradient id="grad1" cx="50%" cy="50%" r="50%" fx="50%" fy="50%">
        <stop offset="0%" stopColor="#60a5fa" stopOpacity="1" />
        <stop offset="100%" stopColor="#1e293b" stopOpacity="1" />
      </radialGradient>
      <radialGradient id="grad2" cx="50%" cy="50%" r="50%" fx="50%" fy="50%">
        <stop offset="0%" stopColor="#fbbf24" stopOpacity="1" />
        <stop offset="100%" stopColor="#1e293b" stopOpacity="1" />
      </radialGradient>
    </defs>
    <ellipse cx="110" cy="110" rx="90" ry="60" fill="url(#grad1)" />
    <ellipse cx="110" cy="110" rx="60" ry="90" fill="url(#grad2)" opacity="0.7" />
    <circle cx="110" cy="110" r="40" fill="#fff" fillOpacity="0.12" />
    <motion.circle
      cx="110"
      cy="60"
      r="18"
      fill="#38bdf8"
      animate={{ cy: [60, 40, 60] }}
      transition={{ repeat: Infinity, duration: 2, ease: 'easeInOut' }}
    />
    <motion.circle
      cx="110"
      cy="160"
      r="12"
      fill="#fbbf24"
      animate={{ cy: [160, 180, 160] }}
      transition={{ repeat: Infinity, duration: 2, ease: 'easeInOut', delay: 1 }}
    />
  </motion.svg>
);

// Dynamically import components with no SSR
const Search = dynamic(() => import('../../components/Search'), {
  ssr: false,
  loading: () => (
    <div className="w-full h-[600px] relative glass rounded-2xl shadow-xl overflow-hidden flex items-center justify-center">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
    </div>
  ),
});

export default function FindPage() {
  return (
    <main className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
      <motion.div
        initial={{ opacity: 0, y: 40 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.8, type: 'spring' }}
        className="container mx-auto px-4 py-8 flex flex-col items-center justify-center min-h-screen"
      >
        {/* Hero Section */}
        <motion.div
          initial={{ scale: 0.95, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ delay: 0.2, duration: 0.7, type: 'spring' }}
          className="text-center mb-12"
        >
          <h1 className="text-4xl md:text-6xl font-extrabold mb-4 bg-gradient-to-r from-blue-400 via-blue-300 to-blue-200 bg-clip-text text-transparent drop-shadow-lg">
            Little Alchemy Recipe Finder
          </h1>
          <p className="text-xl text-blue-100 max-w-2xl mx-auto drop-shadow">
            Visualize how <span className="text-blue-300 font-semibold">BFS</span> and <span className="text-blue-200 font-semibold">DFS</span> algorithms search for element combinations in real-time.
          </p>
        </motion.div>

        {/* Search Section */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5, duration: 0.7, type: 'spring' }}
          className="max-w-4xl w-full mx-auto glass rounded-2xl shadow-2xl border border-white/10 backdrop-blur-md bg-white/5 p-6"
        >
          <Suspense fallback={
            <div className="w-full h-[600px] relative glass rounded-2xl shadow-xl overflow-hidden flex items-center justify-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
            </div>
          }>
            <Search />
          </Suspense>
        </motion.div>

        {/* Footer */}
        <motion.footer
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 1, duration: 0.7 }}
          className="mt-12 text-center text-gray-400"
        >
          <p>Kelompok SOS</p>
        </motion.footer>
      </motion.div>
    </main>
  );
} 