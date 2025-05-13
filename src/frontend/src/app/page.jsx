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
    <rect x="50" y="40" width="80" height="100" rx="24" fill="#e0e7ef" stroke="#60a5fa" strokeWidth="4" />
    <rect x="70" y="20" width="40" height="30" rx="12" fill="#e0e7ef" stroke="#60a5fa" strokeWidth="4" />
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

          {/* Contributors Section */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 1.2, duration: 0.5 }}
            className="mt-32 text-center"
          >
            <h2 className="text-3xl font-bold text-white mb-8"><span className="text-blue-400">AlchemiX</span> Contributors</h2>
            <div className="glass rounded-2xl p-8 border border-white/20 shadow-xl">
              <div className="flex flex-col sm:flex-row items-center justify-center gap-8 sm:gap-16">
                <div className="text-center sm:border-r sm:border-white/20 sm:pr-16">
                  <div className="w-24 h-24 rounded-full bg-blue-500/20 flex items-center justify-center mb-3 overflow-hidden border-2 border-blue-500/30 mx-auto">
                    <img 
                      src="https://github.com/naufal101006.png" 
                      alt="Muhammad Naufal Rayhannida"
                      className="w-full h-full object-cover"
                    />
                  </div>
                  <p className="text-white mb-2">Muhammad Naufal Rayhannida</p>
                  <a 
                    href="https://github.com/naufal101006" 
                    target="_blank" 
                    rel="noopener noreferrer"
                    className="text-blue-400 hover:text-blue-300 transition-colors text-sm flex items-center justify-center gap-1"
                  >
                    <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                    </svg>
                    GitHub Profile
                  </a>
                </div>
                <div className="text-center sm:border-r sm:border-white/20 sm:pr-16">
                  <div className="w-24 h-24 rounded-full bg-blue-500/20 flex items-center justify-center mb-3 overflow-hidden border-2 border-blue-500/30 mx-auto">
                    <img 
                      src="https://github.com/inRiza.png" 
                      alt="Muhammad Rizain Firdaus"
                      className="w-full h-full object-cover"
                    />
                  </div>
                  <p className="text-white mb-2">Muhammad Rizain Firdaus</p>
                  <a 
                    href="https://github.com/inRiza" 
                    target="_blank" 
                    rel="noopener noreferrer"
                    className="text-blue-400 hover:text-blue-300 transition-colors text-sm flex items-center justify-center gap-1"
                  >
                    <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                    </svg>
                    GitHub Profile
                  </a>
                </div>
                <div className="text-center">
                  <div className="w-24 h-24 rounded-full bg-blue-500/20 flex items-center justify-center mb-3 overflow-hidden border-2 border-blue-500/30 mx-auto">
                    <img 
                      src="https://github.com/ivant8k.png" 
                      alt="Ivant Imanuel Silaban"
                      className="w-full h-full object-cover"
                    />
                  </div>
                  <p className="text-white mb-2">Ivant Imanuel Silaban</p>
                  <a 
                    href="https://github.com/ivant8k" 
                    target="_blank" 
                    rel="noopener noreferrer"
                    className="text-blue-400 hover:text-blue-300 transition-colors text-sm flex items-center justify-center gap-1"
                  >
                    <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                    </svg>
                    GitHub Profile
                  </a>
                </div>
              </div>
            </div>
          </motion.div>
        </motion.div>
      </div>
    </main>
  );
} 