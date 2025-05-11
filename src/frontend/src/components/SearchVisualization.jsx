'use client';

import React, { useEffect, useRef, useState } from 'react';
import * as d3 from 'd3';
import { FaPlay, FaPause, FaStepBackward, FaStepForward, FaFastForward, FaUndo } from 'react-icons/fa';

// Helper to build a d3 hierarchy from the solutionPath
function buildTreeFromPath(solutionPath) {
  if (!solutionPath || solutionPath.length === 0) return null;
  const resultNodeMap = new Map();

  // First, create all result nodes
  solutionPath.forEach((step) => {
    if (!resultNodeMap.has(step.result)) {
      resultNodeMap.set(step.result, { name: step.result, children: [] });
    }
  });

  // Then, link children (ingredients are always new nodes)
  solutionPath.forEach((step) => {
    const resultNode = resultNodeMap.get(step.result);
    step.ingredients.forEach((ingredient, i) => {
      // If both ingredients are the same, create two separate nodes
      let ingredientName = ingredient;
      if (step.ingredients[0] === step.ingredients[1]) {
        ingredientName = `${ingredient} (${i + 1})`;
      }
      // If this ingredient is also a result in another step, link to that node
      const ingredientNode = resultNodeMap.has(ingredient)
        ? resultNodeMap.get(ingredient)
        : { name: ingredientName, children: [] };
      resultNode.children.push(ingredientNode);
    });
  });

  // The last step's result is the root
  const root = resultNodeMap.get(solutionPath[solutionPath.length - 1].result);
  return d3.hierarchy(root);
}

const SearchVisualization = ({ element, mode, solutionPath }) => {
  const svgRef = useRef(null);
  const [currentStep, setCurrentStep] = useState(0);
  const [isAnimating, setIsAnimating] = useState(false);
  const [animationSpeed, setAnimationSpeed] = useState(1000); // ms per step

  // Animation interval effect
  useEffect(() => {
    if (!isAnimating) return;
    if (!solutionPath || solutionPath.length === 0) return;
    if (currentStep >= solutionPath.length - 1) {
      setIsAnimating(false);
      return;
    }
    const interval = setInterval(() => {
      setCurrentStep(prev => {
        if (prev >= solutionPath.length - 1) {
          setIsAnimating(false);
          return prev;
        }
        return prev + 1;
      });
    }, animationSpeed);
    return () => clearInterval(interval);
  }, [isAnimating, animationSpeed, currentStep, solutionPath]);

  // Reset animation when solutionPath changes
  useEffect(() => {
    setCurrentStep(0);
    setIsAnimating(false);
  }, [solutionPath]);

  useEffect(() => {
    if (!solutionPath || solutionPath.length === 0) return;
    d3.select(svgRef.current).selectAll("*").remove();

    // Calculate max depth for current step
    const buildTreeUpToStep = (step) => {
      const nodeMap = new Map();
      let root = null;
      for (let i = 0; i <= step; i++) {
        const step = solutionPath[i];
        const result = step.result;
        const [left, right] = step.ingredients;
        if (!nodeMap.has(result)) {
          nodeMap.set(result, { name: result, children: [] });
        }
        if (!nodeMap.has(left)) {
          nodeMap.set(left, { name: left, children: [] });
        }
        if (!nodeMap.has(right)) {
          nodeMap.set(right, { name: right, children: [] });
        }
        const resultNode = nodeMap.get(result);
        const leftNode = nodeMap.get(left);
        const rightNode = nodeMap.get(right);
        resultNode.children = [];
        resultNode.children.push(leftNode);
        resultNode.children.push(rightNode);
        root = resultNode;
      }
      return root;
    };

    // Helper to get max depth
    function getMaxDepth(node, depth = 0) {
      if (!node || !node.children || node.children.length === 0) return depth;
      return Math.max(...node.children.map(child => getMaxDepth(child, depth + 1)));
    }

    const rootData = buildTreeUpToStep(currentStep);
    const maxDepth = getMaxDepth(rootData);
    const baseHeight = window.innerWidth < 640 ? 2000 : 3000; // Smaller height for mobile
    const perLevelHeight = window.innerWidth < 640 ? 500 : 700; // Smaller spacing for mobile
    const height = Math.max(baseHeight, baseHeight + maxDepth * perLevelHeight);
    const width = window.innerWidth < 640 ? 800 : 3000; // Smaller width for mobile
    const nodeRadius = window.innerWidth < 640 ? 40 : 60; // Smaller nodes for mobile
    const margin = { 
      top: window.innerWidth < 640 ? 10 : 20, 
      right: window.innerWidth < 640 ? 45 : 90, 
      bottom: window.innerWidth < 640 ? 40 : 80, 
      left: window.innerWidth < 640 ? 45 : 90 
    };
    const svg = d3.select(svgRef.current)
      .attr('width', width)
      .attr('height', height)
      .attr('viewBox', [0, 0, width, height])
      .attr('style', 'max-width: 100%; height: auto;');
    const zoom = d3.zoom()
      .scaleExtent([0.1, 4])
      .on('zoom', (event) => {
        g.attr('transform', event.transform);
      });
    svg.call(zoom);
    const g = svg.append('g')
      .attr('transform', `translate(${margin.left},${margin.top})`);
    const treeLayout = d3.tree()
      .size([height - margin.top - margin.bottom, width - margin.left - margin.right])
      .separation((a, b) => {
        if (a.depth === b.depth) {
          const siblings = a.parent ? a.parent.children.length : 1;
          return 8 + (siblings * 1.2);
        }
        return 3.5;
      });
    const updateVisualization = (step) => {
      g.selectAll('.link, .node').remove();
      const root = d3.hierarchy(buildTreeUpToStep(step));
      const treeData = treeLayout(root);
      g.selectAll('.link')
        .data(treeData.links())
        .enter()
        .append('path')
        .attr('class', 'link')
        .attr('fill', 'none')
        .attr('stroke', '#22c55e')
        .attr('stroke-width', 2)
        .attr('stroke-opacity', 0.6)
        .attr('d', d3.linkHorizontal()
          .x(d => d.y)
          .y(d => d.x));
      const nodes = g.selectAll('.node')
        .data(treeData.descendants())
        .enter()
        .append('g')
        .attr('class', 'node')
        .attr('transform', d => `translate(${d.y},${d.x})`)
        .style('opacity', 1);
      nodes.append('rect')
        .attr('x', -nodeRadius)
        .attr('y', -nodeRadius / 1.2)
        .attr('width', nodeRadius * 2)
        .attr('height', nodeRadius * 1.7)
        .attr('rx', 14)
        .attr('fill', d => {
          if (d.depth === 0) return '#22c55e';
          if (d.data.children.length === 0) return '#f3f4f6';
          return '#6366f1';
        })
        .attr('stroke', d => {
          if (d.depth === 0) return '#16a34a';
          if (d.data.children.length === 0) return '#9ca3af';
          return '#4f46e5';
        })
        .attr('stroke-width', d => d.depth === 0 ? 3 : 2)
        .attr('stroke-opacity', 0.8);

      // Update text size based on screen width
      const fontSize = window.innerWidth < 640 ? 14 : 20;
      const tierFontSize = window.innerWidth < 640 ? 10 : 12;

      // Add element name
      nodes.append('text')
        .attr('text-anchor', 'middle')
        .attr('dy', '-0.2em')
        .attr('font-size', fontSize)
        .attr('fill', d => {
          if (d.depth === 0) return '#fff';
          if (d.data.children.length === 0) return '#1f2937';
          return '#fff';
        })
        .text(d => d.data.name)
        .style('opacity', 1);

      // Add tier information
      nodes.append('text')
        .attr('text-anchor', 'middle')
        .attr('dy', '1em')
        .attr('font-size', tierFontSize)
        .attr('fill', d => {
          if (d.depth === 0) return '#fff';
          if (d.data.children.length === 0) return '#1f2937';
          return '#fff';
        })
        .text(d => {
          const step = solutionPath.find(s => s.result === d.data.name);
          if (step) {
            return `Tier ${step.tiers.result}`;
          }
          // For leaf nodes, find their tier from the ingredients
          const ingredientStep = solutionPath.find(s => 
            s.ingredients.includes(d.data.name)
          );
          if (ingredientStep) {
            const index = ingredientStep.ingredients.indexOf(d.data.name);
            return `Tier ${index === 0 ? ingredientStep.tiers.left : ingredientStep.tiers.right}`;
          }
          return '';
        })
        .style('opacity', 0.8);

      nodes.append('title')
        .text(d => {
          const step = solutionPath.find(s => s.result === d.data.name);
          if (step) {
            return `${d.data.name} (Tier ${step.tiers.result})`;
          }
          const ingredientStep = solutionPath.find(s => 
            s.ingredients.includes(d.data.name)
          );
          if (ingredientStep) {
            const index = ingredientStep.ingredients.indexOf(d.data.name);
            return `${d.data.name} (Tier ${index === 0 ? ingredientStep.tiers.left : ingredientStep.tiers.right})`;
          }
          return d.data.name;
        });
    };
    updateVisualization(currentStep);
  }, [element, mode, solutionPath, currentStep, animationSpeed]);

  // React-based controls
  const handlePlayPause = () => {
    setIsAnimating((prev) => !prev);
  };
  const handleReset = () => {
    setCurrentStep(0);
    setIsAnimating(false);
  };
  const handleNext = () => {
    setCurrentStep((prev) => Math.min(prev + 1, (solutionPath?.length || 1) - 1));
    setIsAnimating(false);
  };
  const handlePrev = () => {
    setCurrentStep((prev) => Math.max(prev - 1, 0));
    setIsAnimating(false);
  };
  const handleSpeed = () => {
    setAnimationSpeed((prev) => prev === 1000 ? 500 : prev === 500 ? 2000 : 1000);
  };
  const handleEndResult = () => {
    if (solutionPath && solutionPath.length > 0) {
      setCurrentStep(solutionPath.length - 1);
      setIsAnimating(false);
    }
  };

  // Helper function to wrap text
  function wrap(text, width) {
    text.each(function() {
      const text = d3.select(this);
      const words = text.text().split(/\s+/).reverse();
      let word;
      let line = [];
      let lineNumber = 0;
      const lineHeight = 1.1;
      const y = text.attr('y');
      const dy = parseFloat(text.attr('dy'));
      let tspan = text.text(null).append('tspan').attr('x', 0).attr('y', y).attr('dy', dy + 'em');
      
      while (word = words.pop()) {
        line.push(word);
        tspan.text(line.join(' '));
        if (tspan.node().getComputedTextLength() > width) {
          line.pop();
          tspan.text(line.join(' '));
          line = [word];
          tspan = text.append('tspan').attr('x', 0).attr('y', y).attr('dy', ++lineNumber * lineHeight + dy + 'em').text(word);
        }
      }
    });
  }

  const buildTree = (path) => {
    const nodeMap = new Map();
    let root = null;

    // First pass: Create all nodes
    path.forEach(step => {
      const result = step.result;
      const [left, right] = step.ingredients;

      // Create nodes if they don't exist
      if (!nodeMap.has(result)) {
        nodeMap.set(result, { name: result, children: [] });
      }
      if (!nodeMap.has(left)) {
        nodeMap.set(left, { name: left, children: [] });
      }
      if (!nodeMap.has(right)) {
        nodeMap.set(right, { name: right, children: [] });
      }
    });

    // Second pass: Connect nodes
    path.forEach(step => {
      const result = step.result;
      const [left, right] = step.ingredients;

      const resultNode = nodeMap.get(result);
      const leftNode = nodeMap.get(left);
      const rightNode = nodeMap.get(right);

      // Clear existing children to avoid duplicates
      resultNode.children = [];

      // Add children in the correct order
      resultNode.children.push(leftNode);
      resultNode.children.push(rightNode);

      // Set root to the last step's result
      root = resultNode;
    });

    return root;
  };

  return (
    <div className="relative w-full overflow-x-auto">
      <svg ref={svgRef} className="w-full h-full" />
      
      {/* Controls */}
      <div className="absolute top-2 sm:top-4 left-2 sm:left-4 flex items-center gap-1 sm:gap-2 bg-black/50 backdrop-blur-sm p-1.5 sm:p-2 rounded-xl shadow-lg z-10">
        <div className="text-white text-xs px-1.5 sm:px-2 py-0.5 sm:py-1 border-r border-white/20">
          Step {currentStep + 1} of {solutionPath?.length || 0}
        </div>
        <button
          onClick={handleReset}
          className="flex items-center gap-0.5 sm:gap-1 px-1.5 sm:px-2 py-0.5 sm:py-1 rounded-lg bg-white/10 hover:bg-white/20 text-white transition-all duration-200 hover:scale-110 text-xs"
          title="Reset"
        >
          <FaUndo className="w-3 h-3 sm:w-4 sm:h-4" />
          <span className="hidden sm:inline">Reset</span>
        </button>
        
        <button
          onClick={handlePrev}
          className="flex items-center gap-0.5 sm:gap-1 px-1.5 sm:px-2 py-0.5 sm:py-1 rounded-lg bg-white/10 hover:bg-white/20 text-white transition-all duration-200 hover:scale-110 text-xs"
          title="Previous Step"
        >
          <FaStepBackward className="w-3 h-3 sm:w-4 sm:h-4" />
          <span className="hidden sm:inline">Prev</span>
        </button>
        
        <button
          onClick={handlePlayPause}
          className="flex items-center gap-0.5 sm:gap-1 px-1.5 sm:px-2 py-0.5 sm:py-1 rounded-lg bg-blue-500 hover:bg-blue-600 text-white transition-all duration-200 hover:scale-110 text-xs"
          title={isAnimating ? "Pause" : "Play"}
        >
          {isAnimating ? (
            <>
              <FaPause className="w-3 h-3 sm:w-4 sm:h-4" />
              <span className="hidden sm:inline">Pause</span>
            </>
          ) : (
            <>
              <FaPlay className="w-3 h-3 sm:w-4 sm:h-4" />
              <span className="hidden sm:inline">Play</span>
            </>
          )}
        </button>
        
        <button
          onClick={handleNext}
          className="flex items-center gap-0.5 sm:gap-1 px-1.5 sm:px-2 py-0.5 sm:py-1 rounded-lg bg-white/10 hover:bg-white/20 text-white transition-all duration-200 hover:scale-110 text-xs"
          title="Next Step"
        >
          <FaStepForward className="w-3 h-3 sm:w-4 sm:h-4" />
          <span className="hidden sm:inline">Next</span>
        </button>
        
        <button
          onClick={handleEndResult}
          className="flex items-center gap-0.5 sm:gap-1 px-1.5 sm:px-2 py-0.5 sm:py-1 rounded-lg bg-white/10 hover:bg-white/20 text-white transition-all duration-200 hover:scale-110 text-xs"
          title="Show Final Result"
        >
          <FaFastForward className="w-3 h-3 sm:w-4 sm:h-4" />
          <span className="hidden sm:inline">End</span>
        </button>
        
        <button
          onClick={handleSpeed}
          className="flex items-center gap-0.5 sm:gap-1 px-1.5 sm:px-2 py-0.5 sm:py-1 rounded-lg bg-white/10 hover:bg-white/20 text-white transition-all duration-200 hover:scale-110 text-xs"
          title="Change Speed"
        >
          <span>{animationSpeed === 1000 ? '1x' : animationSpeed === 500 ? '2x' : '0.5x'}</span>
        </button>
      </div>
    </div>
  );
};

export default SearchVisualization; 