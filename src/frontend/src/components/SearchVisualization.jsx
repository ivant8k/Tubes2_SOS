'use client';

import React, { useEffect, useRef } from 'react';
import * as d3 from 'd3';

// Helper to build a d3 hierarchy from the solutionPath
function buildTreeFromPath(solutionPath) {
  if (!solutionPath || solutionPath.length === 0) return null;
  // Build a tree where each step's result is a node, and its ingredients are children
  // We'll use a map to avoid duplicate nodes
  const nodeMap = new Map();

  // First, create all nodes
  solutionPath.forEach((step) => {
    if (!nodeMap.has(step.result)) {
      nodeMap.set(step.result, { name: step.result, children: [] });
    }
    step.ingredients.forEach((ingredient) => {
      if (!nodeMap.has(ingredient)) {
        nodeMap.set(ingredient, { name: ingredient, children: [] });
      }
    });
  });

  // Then, link children
  solutionPath.forEach((step) => {
    const resultNode = nodeMap.get(step.result);
    step.ingredients.forEach((ingredient) => {
      const ingredientNode = nodeMap.get(ingredient);
      // Only add as child if not already present
      if (!resultNode.children.includes(ingredientNode)) {
        resultNode.children.push(ingredientNode);
      }
    });
  });

  // The last step's result is the root
  const root = nodeMap.get(solutionPath[solutionPath.length - 1].result);
  return d3.hierarchy(root);
}

const width = 800;
const height = 400;
const nodeRadius = 24;

const SearchVisualization = ({ solutionPath = [] }) => {
  const svgRef = useRef(null);

  useEffect(() => {
    // Clear previous SVG
    d3.select(svgRef.current).selectAll('*').remove();
    if (!solutionPath || solutionPath.length === 0) return;

    const treeData = buildTreeFromPath(solutionPath);
    if (!treeData) return;

    // Create a d3 tree layout (horizontal)
    const treeLayout = d3.tree().size([height - 40, width - 120]);
    const root = treeLayout(treeData);

    const svg = d3.select(svgRef.current)
      .attr('width', width)
      .attr('height', height)
      .attr('viewBox', `0 0 ${width} ${height}`)
      .style('background', '#1e293b');

    // Draw links
    svg.append('g')
      .selectAll('path')
      .data(root.links())
      .join('path')
      .attr('d', d3.linkHorizontal()
        .x(d => d.y + 60)
        .y(d => d.x + 20)
      )
      .attr('fill', 'none')
      .attr('stroke', '#22c55e')
      .attr('stroke-width', 3);

    // Draw nodes
    const nodeG = svg.append('g')
      .selectAll('g')
      .data(root.descendants())
      .join('g')
      .attr('transform', d => `translate(${d.y + 60},${d.x + 20})`);

    nodeG.append('circle')
      .attr('r', nodeRadius)
      .attr('fill', d => d.depth === 0 ? '#22c55e' : '#f3f4f6')
      .attr('stroke', d => d.depth === 0 ? '#16a34a' : '#6366f1')
      .attr('stroke-width', d => d.depth === 0 ? 4 : 2);

    nodeG.append('text')
      .attr('text-anchor', 'middle')
      .attr('dy', '0.35em')
      .attr('font-size', 14)
      .attr('fill', d => d.depth === 0 ? '#fff' : '#1f2937')
      .text(d => d.data.name);
  }, [solutionPath]);

  return (
    <div className="w-full h-[400px] flex items-center justify-center">
      <svg ref={svgRef}></svg>
    </div>
  );
};

export default SearchVisualization; 