import json
import matplotlib.pyplot as plt

def create_benchmark_chart(filename='benches/go_xl-exp-60-sec.json'):
    # Load data
    with open(filename, 'r') as f:
        data = json.load(f)
    
    snaps = data['snaps']
    config = data['conf']
    
    # Extract time series data
    doc_per_sec = [s['doc_per_sec'] for s in snaps]
    miss_per_sec = [s['miss_per_sec'] for s in snaps]
    snapshots = list(range(len(snaps)))
    
    # Create the plot with two subplots
    fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(12, 10))
    
    # Plot 1: Throughput
    ax1.plot(snapshots, doc_per_sec, color='tab:blue', linewidth=2)
    ax1.set_ylabel('Documents per Second', color='tab:blue')
    ax1.set_title('Document Processing Throughput', fontsize=14)
    ax1.tick_params(axis='y', labelcolor='tab:blue')
    ax1.grid(True, alpha=0.3)
    
    # Plot 2: Miss Rate
    ax2.plot(snapshots, miss_per_sec, color='tab:red', linewidth=2)
    ax2.set_ylabel('Misses per Second', color='tab:red')
    ax2.set_xlabel('Snapshot')
    ax2.set_title('Miss Rate Over Time', fontsize=14)
    ax2.tick_params(axis='y', labelcolor='tab:red')
    ax2.grid(True, alpha=0.3)
    
    plt.tight_layout()
    
    # Add info box at bottom
    config_text = f"""Configuration: Batch Size: {config['docs_per_batch']:,}  •  Concurrency: {config['concurrency']}  •  Vector Size: {config['vec_size']}  •  Content Size: {config['content_size']}

Totals: Documents: {data['total_docs']:,}  •  Duration: {data['total_duration'] / 1e9:.1f}s  •  Avg Docs/sec: {data['doc_per_sec']:.0f}  •  Avg Miss/sec: {data['miss_per_sec']:.3f}  •  Total Misses: {data['total_misses']}"""
    
    plt.figtext(0.5, 0.02, config_text, fontsize=11, ha='center',
                bbox=dict(boxstyle="round,pad=0.8", facecolor="white", 
                         edgecolor="gray", alpha=0.9))
    
    plt.subplots_adjust(bottom=0.12)  # Make room for text
    plt.savefig('benchmark_chart.png', dpi=300, bbox_inches='tight')
    plt.show()
    
    print("Chart saved as benchmark_chart.png")

if __name__ == "__main__":
    create_benchmark_chart()
