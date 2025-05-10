'use client';

import React, { useEffect, useRef, useState } from 'react';
import * as d3 from 'd3';

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
    const width = 1200;
    const height = 600;
    const nodeRadius = 24;
    const margin = { top: 20, right: 90, bottom: 30, left: 90 };
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
          return 4 + (siblings * 0.8);
        }
        return 1.1;
      });
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
      nodes.append('circle')
        .attr('r', nodeRadius)
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
      nodes.append('text')
        .attr('text-anchor', 'middle')
        .attr('dy', '0.35em')
        .attr('font-size', 14)
        .attr('fill', d => {
          if (d.depth === 0) return '#fff';
          if (d.data.children.length === 0) return '#1f2937';
          return '#fff';
        })
        .text(d => d.data.name)
        .style('opacity', 1);
      nodes.append('title')
        .text(d => d.data.name);
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
    <div className="w-full overflow-hidden">
      <div className="flex items-center gap-2 mb-2">
        <button onClick={handlePrev} className="px-3 py-1 bg-slate-700 text-white rounded disabled:opacity-50" disabled={currentStep === 0}>
          Prev
        </button>
        <button onClick={handlePlayPause} className="px-3 py-1 bg-green-600 text-white rounded" disabled={!solutionPath || solutionPath.length === 0}>
          {isAnimating ? 'Pause' : 'Play'}
        </button>
        <button onClick={handleNext} className="px-3 py-1 bg-slate-700 text-white rounded disabled:opacity-50" disabled={!solutionPath || currentStep >= (solutionPath?.length || 1) - 1}>
          Next
        </button>
        <button onClick={handleReset} className="px-3 py-1 bg-blue-600 text-white rounded">
          Reset
        </button>
        <button onClick={handleSpeed} className="px-3 py-1 bg-slate-800 text-white rounded">
          Speed: {animationSpeed}ms
        </button>
        <span className="ml-4 text-slate-700 dark:text-slate-200">Step {solutionPath ? currentStep + 1 : 0}/{solutionPath ? solutionPath.length : 0}</span>
      </div>
      <svg ref={svgRef} className="w-full h-[600px]"></svg>
    </div>
  );
};

export default SearchVisualization; 