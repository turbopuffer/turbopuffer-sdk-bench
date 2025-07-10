import json
import matplotlib.pyplot as plt
import glob
import os

def load_benchmark_files(directory='benches'):
    """Load all benchmark JSON files from the specified directory"""
    pattern = os.path.join(directory, '*.json')
    files = glob.glob(pattern)
    benchmarks = []
    
    for file in files:
        try:
            with open(file, 'r') as f:
                data = json.load(f)
                benchmarks.append({
                    'filename': os.path.basename(file),
                    'data': data,
                    'config': data['conf'],
                    'avg_docs_per_sec': data['doc_per_sec'],
                    'avg_miss_per_sec': data['miss_per_sec'],
                    'total_misses': data['total_misses']
                })
        except Exception as e:
            print(f"Error loading {file}: {e}")
    
    return benchmarks

def create_comparison_chart(benchmarks):
    """Create single chart with three lines from benchmark totals"""
    
    if not benchmarks:
        print("No benchmarks found")
        return
    
    # Sort by size order (sm, md, lg, xl)
    size_order = {'sm': 0, 'md': 1, 'lg': 2, 'xl': 3}
    benchmarks.sort(key=lambda x: size_order.get(x['filename'].replace('.json', ''), 999))
    
    # Extract data for plotting
    labels = [b['filename'].replace('.json', '') for b in benchmarks]
    avg_docs_per_sec = [b['avg_docs_per_sec'] for b in benchmarks]
    avg_miss_per_sec = [b['avg_miss_per_sec'] for b in benchmarks]
    total_misses = [b['total_misses'] for b in benchmarks]
    
    x_positions = range(len(labels))
    
    # Create the plot
    fig, ax = plt.subplots(figsize=(12, 8))
    
    # Plot all three lines
    ax.plot(x_positions, avg_docs_per_sec, 'b-o', linewidth=2, markersize=8, label='Avg Docs/sec')
    ax.plot(x_positions, avg_miss_per_sec, 'r-s', linewidth=2, markersize=8, label='Avg Miss/sec')
    ax.plot(x_positions, total_misses, 'g-^', linewidth=2, markersize=8, label='Total Misses')
    
    # Set y-axis limits to accommodate all data points and labels
    max_value = max(max(avg_docs_per_sec), max(avg_miss_per_sec), max(total_misses))
    min_value = min(min(avg_docs_per_sec), min(avg_miss_per_sec), min(total_misses))
    
    # Add padding: 15% at top for labels, 10% at bottom for low values
    y_range = max_value - min_value
    ax.set_ylim(min_value - (y_range * 0.1), max_value + (y_range * 0.15))
    
    ax.set_xlabel('Benchmark')
    ax.set_ylabel('Value')
    ax.set_title('Benchmark Performance Comparison')
    ax.set_xticks(x_positions)
    ax.set_xticklabels(labels)
    ax.legend()
    ax.grid(True, alpha=0.3)
    
    # Add value labels on points with better positioning
    for i, (docs, miss, total) in enumerate(zip(avg_docs_per_sec, avg_miss_per_sec, total_misses)):
        ax.annotate(f'{docs:.0f}', (i, docs), textcoords="offset points", 
                   xytext=(0,15), ha='center', fontsize=9, color='blue')
        # Only show miss rate labels if they're not zero
        if miss > 0:
            ax.annotate(f'{miss:.3f}', (i, miss), textcoords="offset points", 
                       xytext=(0,15), ha='center', fontsize=9, color='red')
        # Only show total misses labels if they're not zero
        if total > 0:
            ax.annotate(f'{total}', (i, total), textcoords="offset points", 
                       xytext=(0,15), ha='center', fontsize=9, color='green')
    
    plt.tight_layout()
    
    # Add benchmark details at bottom as a table
    table_header = f"{'Benchmark':<12} {'Batch':<10} {'Concurrency':<12} {'Total Docs':<12} {'Duration':<10} {'Throughput':<12} {'Miss Rate':<12} {'Total Misses':<12}"
    table_divider = "-" * 110
    
    details_text = table_header + "\n" + table_divider + "\n"
    
    for i, b in enumerate(benchmarks):
        config = b['config']
        name = labels[i].upper()
        batch = f"{config['docs_per_batch']:,}"
        concurrency = str(config['concurrency'])
        total_docs = f"{b['data']['total_docs']:,}"
        duration = f"{b['data']['total_duration'] / 1e9:.1f}s"
        throughput = f"{b['avg_docs_per_sec']:.0f}/s"
        miss_rate = f"{b['avg_miss_per_sec']:.3f}/s"
        total_misses = str(b['total_misses'])
        
        row = f"{name:<12} {batch:<10} {concurrency:<12} {total_docs:<12} {duration:<10} {throughput:<12} {miss_rate:<12} {total_misses:<12}"
        details_text += row
        if i < len(benchmarks) - 1:  # Add newline except for last row
            details_text += "\n"
    
    plt.figtext(0.5, 0.02, details_text, fontsize=9, ha='center', 
                fontfamily='monospace',
                bbox=dict(boxstyle="round,pad=1.0", facecolor="lightgray", 
                         edgecolor="darkgray", alpha=0.9))
    
    plt.subplots_adjust(bottom=0.25)  # Back to original bottom margin
    
    # Save chart
    filename = os.path.join('benches', 'benchmark_comparison.png')
    plt.savefig(filename, dpi=300, bbox_inches='tight')
    plt.show()
    
    print(f"Comparison chart saved as {filename}")

def main():
    """Main function"""
    print("Loading benchmark files from benches/ directory...")
    
    # Load all benchmark files
    benchmarks = load_benchmark_files()
    
    if not benchmarks:
        print("No JSON files found in benches/ directory.")
        return
    
    print(f"Found {len(benchmarks)} benchmark files")
    for b in benchmarks:
        print(f"  {b['filename']}: {b['avg_docs_per_sec']:.0f} docs/sec, {b['total_misses']} misses")
    
    # Create comparison chart
    create_comparison_chart(benchmarks)

if __name__ == "__main__":
    main()
