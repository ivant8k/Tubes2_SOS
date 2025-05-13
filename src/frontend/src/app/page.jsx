"use client";

import React from 'react';
import { motion } from 'framer-motion';
import { useRouter } from 'next/navigation';

const Beaker3D = () => (
  <motion.svg
    width="180"
    height="180"
    viewBox="0 0 180 180"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    initial={{ rotate: -8, scale: 0.9 }}
    animate={{ rotate: 8, scale: 1 }}
    transition={{ yoyo: Infinity, duration: 2, ease: 'easeInOut' }}
    className="drop-shadow-2xl"
  >
    {/* Beaker body */}
    <rect x="50" y="40" width="80" height="100" rx="24" fill="#e0e7ef" stroke="#60a5fa" strokeWidth="4" />
    {/* Beaker neck */}
    <rect x="70" y="20" width="40" height="30" rx="12" fill="#e0e7ef" stroke="#60a5fa" strokeWidth="4" />
    {/* Liquid */}
    <motion.ellipse
      cx="90"
      cy="110"
      rx="32"
      ry="18"
      fill="#38bdf8"
      initial={{ cy: 120 }}
      animate={{ cy: [120, 110, 120] }}
      transition={{ repeat: Infinity, duration: 2, ease: 'easeInOut' }}
      opacity={0.85}
    />
    {/* Bubbles */}
    <motion.circle
      cx="90"
      cy="90"
      r="6"
      fill="#bae6fd"
      animate={{ cy: [90, 70, 90] }}
      transition={{ repeat: Infinity, duration: 2, ease: 'easeInOut', delay: 0.5 }}
      opacity={0.7}
    />
    <motion.circle
      cx="110"
      cy="100"
      r="3"
      fill="#bae6fd"
      animate={{ cy: [100, 80, 100] }}
      transition={{ repeat: Infinity, duration: 2, ease: 'easeInOut', delay: 1 }}
      opacity={0.7}
    />
    <motion.circle
      cx="75"
      cy="105"
      r="2.5"
      fill="#bae6fd"
      animate={{ cy: [105, 95, 105] }}
      transition={{ repeat: Infinity, duration: 2, ease: 'easeInOut', delay: 1.3 }}
      opacity={0.7}
    />
    {/* Beaker shine */}
    <ellipse cx="80" cy="70" rx="8" ry="18" fill="#fff" opacity="0.18" />
  </motion.svg>
);

export default function LandingPage() {
  const router = useRouter();
  return (
    <main className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
      <div className="container mx-auto px-4 py-8 flex flex-col items-center justify-center min-h-screen">
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 1 }}
          className="flex flex-col items-center gap-8"
        >
          <Beaker3D />
          <motion.h1
            initial={{ scale: 0.8, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            transition={{ delay: 0.3, duration: 0.7, type: 'spring' }}
            className="text-4xl sm:text-6xl font-extrabold text-white drop-shadow-lg text-center"
          >
            Welcome to <span className="text-blue-400">AlchemiX</span>
          </motion.h1>
          <motion.p
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.6, duration: 0.7 }}
            className="text-lg sm:text-xl text-white max-w-xl text-center"
          >
            Discover, combine, and visualize magical elements. Start your journey to find the perfect recipe!
          </motion.p>
          <motion.button
            whileHover={{
              scale: 1.08,
              backgroundColor: 'rgba(56,189,248,0.25)',
              color: '#fff',
              boxShadow: '0 8px 32px 0 rgba(31, 38, 135, 0.37)',
              backdropFilter: 'blur(8px)',
              border: '1.5px solid rgba(255,255,255,0.18)',
              transition: { duration: 0.1, type: 'spring' }
            }}
            whileTap={{ scale: 0.97 }}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 1, duration: 0.5 }}
            className="mt-4 px-8 py-3 rounded-xl bg-white/10 text-white font-semibold text-lg shadow-xl glass border border-white/20 hover:bg-blue-400/30 hover:text-blue-100 focus:outline-none focus:ring-2 focus:ring-blue-300 transition-all duration-100 backdrop-blur-md"
            style={{
              boxShadow: '0 4px 24px 0 rgba(31, 38, 135, 0.17)',
              backdropFilter: 'blur(6px)',
              border: '1.5px solid rgba(255,255,255,0.12)',
            }}
            onClick={() => router.push('/find')}
          >
            Find my recipe
          </motion.button>
        </motion.div>
      </div>
    </main>
  );
} 