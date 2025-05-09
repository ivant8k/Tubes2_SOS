import { Inter } from 'next/font/google';
import './globals.css';

const inter = Inter({ subsets: ['latin'] });

export const metadata = {
  title: 'Element Search Visualization',
  description: 'Visualize element search algorithms in real-time',
  keywords: 'element search, visualization, algorithm, BFS, DFS',
};

export default function RootLayout({ children }) {
  return (
    <html lang="en" className="dark">
      <body 
        className={`${inter.className} min-h-screen bg-gray-900 text-white antialiased`}
        suppressHydrationWarning
      >
        {children}
      </body>
    </html>
  );
} 