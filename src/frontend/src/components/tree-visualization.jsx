'use client';

import React, { useEffect, useState, useCallback } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  useNodesState,
  useEdgesState,
  Position,
  MarkerType,
  BaseEdge,
  getStraightPath,
  Handle,
} from '@xyflow/react';
import { motion, AnimatePresence } from 'framer-motion';
import '@xyflow/react/dist/style.css';
import { FaPlay, FaPause, FaStepBackward, FaStepForward, FaFastForward, FaUndo } from 'react-icons/fa';

// Custom node component with animation
const CustomNode = ({ data }) => {
  // Modern, minimal, dark theme node style
  let bgColor = 'bg-gray-800';
  let borderColor = 'border-gray-600';
  let textColor = 'text-white';
  if (data.isRoot) {
    bgColor = 'bg-gray-800';
    borderColor = 'border-green-500';
    textColor = 'text-green-300';
  } else if (data.isLeaf) {
    bgColor = 'bg-gray-900';
    borderColor = 'border-gray-700';
    textColor = 'text-gray-200';
  } else {
    bgColor = 'bg-gray-700';
    borderColor = 'border-blue-700';
    textColor = 'text-blue-200';
  }
  return (
    <motion.div
      initial={{ scale: 0, opacity: 0 }}
      animate={{ scale: 1, opacity: 1 }}
      transition={{ duration: 0.5, type: "spring" }}
      className={`px-4 py-2 rounded-xl shadow-md border ${bgColor} ${borderColor} ${textColor} font-medium`}
      style={{
        minWidth: '120px',
        textAlign: 'center',
        borderWidth: 2,
      }}
    >
      {/* Source handle for outgoing edges (ATAS) */}
      <Handle 
        type="source" 
        position={Position.Top} 
        id="top"
        style={{ 
          background: '#3b82f6',
          width: 8,
          height: 8,
          top: -4
        }}
      />
      <div className="font-semibold text-base">{data.label}</div>
      <div className="text-xs opacity-80">Tier {data.tier}</div>
      {/* Target handle for incoming edges (BAWAH) */}
      <Handle 
        type="target" 
        position={Position.Bottom} 
        id="bottom"
        style={{ 
          background: '#3b82f6',
          width: 8,
          height: 8,
          bottom: -4
        }}
      />
    </motion.div>
  );
};

// Custom edge component
const CustomEdge = ({ id, sourceX, sourceY, targetX, targetY, sourceHandleId, targetHandleId }) => {
  const [edgePath] = getStraightPath({
    sourceX,
    sourceY,
    targetX,
    targetY,
  });

  return (
    <BaseEdge
      path={edgePath}
      style={{ stroke: '#3b82f6', strokeWidth: 2 }}
      markerEnd={{
        type: MarkerType.ArrowClosed,
        color: '#3b82f6',
        width: 16,
        height: 16,
      }}
    />
  );
};

const nodeTypes = {
  custom: CustomNode,
};

const edgeTypes = {
  custom: CustomEdge,
};

const defaultEdgeOptions = {
  type: 'smoothstep',
  animated: false,
  style: { stroke: 'blue', strokeWidth: 4 },
  markerEnd: {
    type: MarkerType.ArrowClosed,
    color: 'blue',
    width: 16,
    height: 16,
  },
};

const TreeVisualization = ({ element, mode, solutionPath }) => {
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [currentStep, setCurrentStep] = useState(0);
  const [isAnimating, setIsAnimating] = useState(false);
  const [animationSpeed, setAnimationSpeed] = useState(1000);

  // Replace the buildGraph function with a tree layout
  const buildGraph = useCallback((step) => {
    if (!solutionPath || solutionPath.length === 0) return { nodes: [], edges: [] };

    // Helper to get node key for a level occurrence
    function getNodeKey(element, tier, depth, idx) {
      return `${element}-${tier}-${depth}-${idx}`;
    }

    // Build a map from result to its step index for quick lookup
    const resultStepMap = {};
    for (let i = 0; i <= step; i++) {
      resultStepMap[solutionPath[i].result] = i;
    }

    // --- BFS Layout (default, horizontal spread per level) ---
    function buildBFS() {
      const nodeMap = new Map();
      const edgeList = [];
      let globalX = 0;
      function traverse(result, tier, depth) {
        const i = resultStepMap[result];
        let children = [];
        if (i !== undefined) {
          const { ingredients, tiers } = solutionPath[i];
          const [left, right] = ingredients;
          const leftTier = tiers.left;
          const rightTier = tiers.right;
          const leftKey = traverse(left, leftTier, depth + 1);
          const rightKey = traverse(right, rightTier, depth + 1);
          children = [leftKey, rightKey];
        }
        const idx = Array.from(nodeMap.values()).filter(n => n.label === result && n.tier === tier && n.depth === depth).length;
        const nodeKey = getNodeKey(result, tier, depth, idx);
        if (nodeMap.has(nodeKey)) return nodeKey;
        let x;
        if (children.length === 0) {
          x = globalX++;
        } else {
          const childXs = children.map(childKey => nodeMap.get(childKey).x);
          x = (Math.min(...childXs) + Math.max(...childXs)) / 2;
        }
        nodeMap.set(nodeKey, {
          id: nodeKey,
          label: result,
          tier: tier,
          depth: depth,
          x,
          y: depth,
          isRoot: depth === 0,
          isLeaf: children.length === 0,
        });
        children.forEach(childKey => {
          edgeList.push({
            id: `e-${childKey}-${nodeKey}`,
            source: childKey,
            target: nodeKey,
            type: 'custom',
            animated: false,
            style: { stroke: '#3b82f6', strokeWidth: 2 },
            markerEnd: {
              type: MarkerType.ArrowClosed,
              color: '#3b82f6',
              width: 16,
              height: 16,
            },
            sourceHandle: 'top',
            targetHandle: 'bottom',
          });
        });
        return nodeKey;
      }
      const lastStep = solutionPath[step];
      traverse(lastStep.result, lastStep.tiers.result, 0);
      return { nodeMap, edgeList };
    }

    // --- DFS Layout (vertical/deep, child directly below parent) ---
    function buildDFS() {
      const nodeMap = new Map();
      const edgeList = [];
      let globalX = 0;
      function traverse(result, tier, depth) {
        const i = resultStepMap[result];
        let children = [];
        if (i !== undefined) {
          const { ingredients, tiers } = solutionPath[i];
          const [left, right] = ingredients;
          const leftTier = tiers.left;
          const rightTier = tiers.right;
          const leftKey = traverse(left, leftTier, depth + 1);
          const rightKey = traverse(right, rightTier, depth + 1);
          children = [leftKey, rightKey];
        }
        const idx = Array.from(nodeMap.values()).filter(n => n.label === result && n.tier === tier && n.depth === depth).length;
        const nodeKey = getNodeKey(result, tier, depth, idx);
        if (nodeMap.has(nodeKey)) return nodeKey;
        let x;
        if (children.length === 0) {
          x = globalX++;
        } else {
          const childXs = children.map(childKey => nodeMap.get(childKey).x);
          x = (Math.min(...childXs) + Math.max(...childXs)) / 2;
        }
        nodeMap.set(nodeKey, {
          id: nodeKey,
          label: result,
          tier: tier,
          depth: depth,
          x,
          y: depth,
          isRoot: depth === 0,
          isLeaf: children.length === 0,
        });
        children.forEach(childKey => {
          edgeList.push({
            id: `e-${childKey}-${nodeKey}`,
            source: childKey,
            target: nodeKey,
            type: 'custom',
            animated: false,
            style: { stroke: '#3b82f6', strokeWidth: 2 },
            markerEnd: {
              type: MarkerType.ArrowClosed,
              color: '#3b82f6',
              width: 16,
              height: 16,
            },
            sourceHandle: 'top',
            targetHandle: 'bottom',
          });
        });
        return nodeKey;
      }
      const lastStep = solutionPath[step];
      traverse(lastStep.result, lastStep.tiers.result, 0);
      return { nodeMap, edgeList };
    }

    // --- Bidirectional Layout (two trees meeting at a frontier node) ---
    function buildBidirectional() {
      // For simplicity, treat as BFS for now, but can be extended to two trees
      return buildBFS();
    }

    // --- Select layout based on mode ---
    let nodeMap, edgeList;
    if (mode === 'dfs') {
      ({ nodeMap, edgeList } = buildDFS());
    } else if (mode === 'bidirectional') {
      ({ nodeMap, edgeList } = buildBidirectional());
    } else {
      ({ nodeMap, edgeList } = buildBFS());
    }
    const nodes = Array.from(nodeMap.values()).map(node => ({
      id: node.id,
      type: 'custom',
      position: {
        x: node.x * 180,
        y: node.y * 120,
      },
      data: {
        label: node.label,
        tier: node.tier,
        isRoot: node.isRoot,
        isLeaf: node.isLeaf,
      },
      sourcePosition: Position.Top,
      targetPosition: Position.Bottom,
    }));
    const edges = edgeList.map(e => ({
      id: e.id,
      source: e.source,
      target: e.target,
      type: 'custom',
      animated: false,
      style: { stroke: '#3b82f6', strokeWidth: 2 },
      markerEnd: {
        type: MarkerType.ArrowClosed,
        color: '#3b82f6',
        width: 16,
        height: 16,
      },
    }));
    return { nodes, edges };
  }, [solutionPath, mode]);

  // Animation effect
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

  // Helper to get nodes/edges up to currentStep, step-by-step by solutionPath
  function getStepwiseTree(solutionPath, buildGraph, currentStep) {
    // Always build the full tree for layout
    const { nodes: fullNodes, edges: fullEdges } = buildGraph(solutionPath.length - 1);
    // Accumulate nodes/edges by traversing solutionPath steps
    const unlockedNodeIds = new Set();
    const unlockedEdgeIds = new Set();
    
    // Keep track of which nodes are used as ingredients
    const ingredientNodeMap = new Map();
    
    // For step 0, only show basic elements (tier 0)
    if (currentStep === 0) {
      fullNodes.forEach(node => {
        if (node.data.tier === 0) {
          unlockedNodeIds.add(node.id);
        }
      });
      return { 
        nodes: fullNodes.filter(n => unlockedNodeIds.has(n.id)), 
        edges: [] 
      };
    }
    
    for (let i = 0; i <= currentStep; i++) {
      const step = solutionPath[i];
      if (!step) continue;
      
      // Add result node
      const resultNode = fullNodes.find(n => n.data.label === step.result);
      if (resultNode) {
        unlockedNodeIds.add(resultNode.id);
        
        // Add all edges that connect to this result node
        fullEdges.forEach(edge => {
          if (edge.target === resultNode.id) {
            unlockedEdgeIds.add(edge.id);
            // Also add the source node of this edge
            unlockedNodeIds.add(edge.source);
          }
        });
      }
      
      // Add ingredient nodes and track their usage
      step.ingredients.forEach((ingredient, idx) => {
        // Find all nodes with this ingredient label
        const ingredientNodes = fullNodes.filter(n => n.data.label === ingredient);
        
        // For each ingredient node, check if it's already been used
        ingredientNodes.forEach(node => {
          const key = `${node.id}-${i}`; // Unique key for this step
          if (!ingredientNodeMap.has(key)) {
            ingredientNodeMap.set(key, true);
            unlockedNodeIds.add(node.id);
            
            // Add all edges from this ingredient node
            fullEdges.forEach(edge => {
              if (edge.source === node.id) {
                unlockedEdgeIds.add(edge.id);
                // Also add the target node of this edge
                unlockedNodeIds.add(edge.target);
              }
            });
          }
        });
      });
    }
    
    // Filter full tree to only show unlocked nodes/edges
    const visibleNodes = fullNodes.filter(n => unlockedNodeIds.has(n.id));
    const visibleEdges = fullEdges.filter(e => unlockedEdgeIds.has(e.id));
    
    // Double check that all visible nodes have their corresponding edges
    visibleNodes.forEach(node => {
      fullEdges.forEach(edge => {
        if ((edge.source === node.id || edge.target === node.id) && 
            unlockedNodeIds.has(edge.source) && 
            unlockedNodeIds.has(edge.target)) {
          unlockedEdgeIds.add(edge.id);
        }
      });
    });
    
    return { 
      nodes: visibleNodes, 
      edges: fullEdges.filter(e => unlockedEdgeIds.has(e.id))
    };
  }

  // Update visualization when step changes (step-by-step by solutionPath)
  useEffect(() => {
    const { nodes: visibleNodes, edges: visibleEdges } = getStepwiseTree(solutionPath, buildGraph, currentStep);
    setNodes(visibleNodes);
    setEdges(visibleEdges);
  }, [currentStep, buildGraph, setNodes, setEdges, solutionPath]);

  // Reset animation when solutionPath changes
  useEffect(() => {
    setCurrentStep(0);
    setIsAnimating(false);
  }, [solutionPath]);

  // Control handlers
  const handlePlayPause = () => setIsAnimating(prev => !prev);
  const handleReset = () => {
    setCurrentStep(0);
    setIsAnimating(false);
  };
  const handleNext = () => {
    setCurrentStep(prev => Math.min(prev + 1, (solutionPath?.length || 1) - 1));
    setIsAnimating(false);
  };
  const handlePrev = () => {
    setCurrentStep(prev => Math.max(prev - 1, 0));
    setIsAnimating(false);
  };
  const handleSpeed = () => {
    setAnimationSpeed(prev => prev === 1000 ? 500 : prev === 500 ? 2000 : 1000);
  };
  const handleEndResult = () => {
    if (solutionPath && solutionPath.length > 0) {
      setCurrentStep(solutionPath.length - 1);
      setIsAnimating(false);
    }
  };

  return (
    <div className="relative w-full h-[600px]">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        defaultEdgeOptions={defaultEdgeOptions}
        fitView
        style={{ background: 'transparent' }}
        proOptions={{ hideAttribution: true }}
        minZoom={0.1}
      >
        <Background color="#aaa" gap={16} />
      </ReactFlow>
      
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

export default TreeVisualization; 