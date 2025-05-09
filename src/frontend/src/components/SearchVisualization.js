import React, { useEffect, useRef, useState } from 'react';
import { Network } from 'vis-network';
import { DataSet } from 'vis-data';

const SearchVisualization = ({ element, mode }) => {
  const networkRef = useRef(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const network = useRef(null);
  const nodes = useRef(new DataSet());
  const edges = useRef(new DataSet());

  useEffect(() => {
    if (!networkRef.current) return;

    const options = {
      nodes: {
        shape: 'dot',
        size: 20,
        font: {
          size: 14,
          color: '#ffffff',
        },
        color: {
          background: '#4a90e2',
          border: '#2171c7',
          highlight: {
            background: '#2171c7',
            border: '#1a5ca8',
          },
        },
      },
      edges: {
        width: 2,
        color: {
          color: '#999',
          highlight: '#666',
        },
        smooth: {
          type: 'continuous',
        },
      },
      physics: {
        stabilization: false,
        barnesHut: {
          gravitationalConstant: -80000,
          springConstant: 0.001,
          springLength: 200,
        },
      },
      layout: {
        hierarchical: {
          direction: 'UD',
          sortMethod: 'directed',
          levelSeparation: 150,
          nodeSpacing: 100,
        },
      },
    };

    network.current = new Network(
      networkRef.current,
      { nodes: nodes.current, edges: edges.current },
      options
    );

    const eventSource = new EventSource(
      `http://localhost:5000/visualize?element=${element}&mode=${mode}`
    );

    setIsLoading(true);
    setError(null);

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        
        // Update nodes
        data.nodes.forEach((node) => {
          if (!nodes.current.get(node.id)) {
            nodes.current.add({
              id: node.id,
              label: node.label,
              level: node.depth,
              color: {
                background: '#4a90e2',
                border: '#2171c7',
                highlight: {
                  background: '#2171c7',
                  border: '#1a5ca8',
                },
              },
            });
          }
        });

        // Update edges
        data.edges.forEach((edge) => {
          if (!edges.current.get(`${edge.from}-${edge.to}`)) {
            edges.current.add({
              from: edge.from,
              to: edge.to,
              id: `${edge.from}-${edge.to}`,
            });
          }
        });

        // Highlight the last node (currently being visited)
        if (data.nodes.length > 0) {
          const lastNode = data.nodes[data.nodes.length - 1];
          nodes.current.update({
            id: lastNode.id,
            color: {
              background: '#e24a4a',
              border: '#c72121',
              highlight: {
                background: '#c72121',
                border: '#a81a1a',
              },
            },
          });
        }

        // Fit the network to show all nodes
        network.current?.fit();
      } catch (err) {
        console.error('Error processing visualization data:', err);
        setError('Error processing visualization data');
      }
    };

    eventSource.onerror = (err) => {
      console.error('EventSource error:', err);
      setError('Error connecting to visualization stream');
      eventSource.close();
      setIsLoading(false);
    };

    eventSource.addEventListener('end', () => {
      eventSource.close();
      setIsLoading(false);
    });

    return () => {
      eventSource.close();
      if (network.current) {
        network.current.destroy();
      }
      nodes.current.clear();
      edges.current.clear();
    };
  }, [element, mode]);

  return (
    <div className="w-full h-[600px] relative">
      {isLoading && (
        <div className="absolute inset-0 flex items-center justify-center bg-white bg-opacity-75 z-10">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
        </div>
      )}
      {error && (
        <div className="absolute inset-0 flex items-center justify-center bg-white bg-opacity-75 z-10">
          <div className="text-red-500">{error}</div>
        </div>
      )}
      <div ref={networkRef} className="w-full h-full" />
    </div>
  );
};

export default SearchVisualization; 