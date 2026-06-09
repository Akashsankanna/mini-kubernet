#!/usr/bin/env python3

"""
AI-Based Auto Scaler for Kubernetes
Uses machine learning to predict metrics and scale pods proactively
"""

import os
import json
import time
import logging
from datetime import datetime, timedelta
import numpy as np
from kubernetes import client, config, watch
import requests

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class MetricsPredictor:
    """Predict future metrics using ML models"""
    
    def __init__(self, prometheus_url):
        self.prometheus_url = prometheus_url
    
    def get_metrics(self, query, start_time, end_time):
        """Fetch metrics from Prometheus"""
        url = f"{self.prometheus_url}/api/v1/query_range"
        params = {
            'query': query,
            'start': int(start_time.timestamp()),
            'end': int(end_time.timestamp()),
            'step': '30s'
        }
        
        try:
            response = requests.get(url, params=params, timeout=10)
            response.raise_for_status()
            return response.json()['data']['result']
        except Exception as e:
            logger.error(f"Error fetching metrics: {e}")
            return []
    
    def predict_cpu(self, container_name, namespace):
        """Predict CPU usage for next 30 minutes"""
        query = f'rate(container_cpu_usage_seconds_total{{pod="{container_name}",namespace="{namespace}"}}[5m])'
        
        end_time = datetime.now()
        start_time = end_time - timedelta(hours=24)
        
        metrics = self.get_metrics(query, start_time, end_time)
        
        if not metrics:
            return None
        
        # Extract values
        values = [float(v[1]) for v in metrics[0]['values'] if v[1] != 'NaN']
        
        if len(values) < 10:
            return np.mean(values) if values else None
        
        # Simple LSTM-like prediction using exponential smoothing
        alpha = 0.3
        prediction = values[-1]
        for v in reversed(values[-20:]):
            prediction = alpha * v + (1 - alpha) * prediction
        
        return prediction
    
    def predict_memory(self, container_name, namespace):
        """Predict memory usage for next 30 minutes"""
        query = f'container_memory_usage_bytes{{pod="{container_name}",namespace="{namespace}"}}'
        
        end_time = datetime.now()
        start_time = end_time - timedelta(hours=24)
        
        metrics = self.get_metrics(query, start_time, end_time)
        
        if not metrics:
            return None
        
        values = [float(v[1]) for v in metrics[0]['values'] if v[1] != 'NaN']
        
        if not values:
            return None
        
        # Use moving average
        return np.mean(values[-100:]) if len(values) > 100 else np.mean(values)

class AIScaler:
    """AI-based Kubernetes autoscaler"""
    
    def __init__(self, namespace, config_map_name):
        config.load_incluster_config()
        self.v1 = client.CoreV1Api()
        self.apps_v1 = client.AppsV1Api()
        self.namespace = namespace
        self.config_map_name = config_map_name
        self.predictor = MetricsPredictor(
            os.getenv('PROMETHEUS_URL', 'http://prometheus:9090')
        )
    
    def load_config(self):
        """Load scaling configuration from ConfigMap"""
        try:
            cm = self.v1.read_namespaced_config_map(
                self.config_map_name, self.namespace
            )
            return json.loads(cm.data['config.yaml'])
        except Exception as e:
            logger.error(f"Error loading config: {e}")
            return {}
    
    def get_deployment_replicas(self, deployment_name):
        """Get current replica count"""
        try:
            deployment = self.apps_v1.read_namespaced_deployment(
                deployment_name, self.namespace
            )
            return deployment.spec.replicas
        except Exception as e:
            logger.error(f"Error getting deployment: {e}")
            return None
    
    def scale_deployment(self, deployment_name, replicas):
        """Scale deployment to desired replica count"""
        try:
            deployment = self.apps_v1.read_namespaced_deployment(
                deployment_name, self.namespace
            )
            deployment.spec.replicas = replicas
            self.apps_v1.patch_namespaced_deployment(
                deployment_name, self.namespace, deployment
            )
            logger.info(f"Scaled {deployment_name} to {replicas} replicas")
        except Exception as e:
            logger.error(f"Error scaling deployment: {e}")
    
    def calculate_desired_replicas(self, deployment_name, config):
        """Calculate desired replica count based on predictions"""
        policy = config.get('scaling_policies', {}).get(deployment_name, {})
        
        if not policy:
            return None
        
        min_replicas = policy.get('min_replicas', 1)
        max_replicas = policy.get('max_replicas', 10)
        
        current_replicas = self.get_deployment_replicas(deployment_name)
        if current_replicas is None:
            return None
        
        # Get predicted metrics
        metrics = policy.get('metrics', [])
        total_weight = 0
        weighted_score = 0
        
        for metric in metrics:
            metric_name = metric.get('name')
            target = metric.get('target')
            weight = metric.get('prediction_weight', 1.0)
            
            # Get predicted value
            if metric_name == 'cpu':
                predicted = self.predictor.predict_cpu(deployment_name, self.namespace)
            elif metric_name == 'memory':
                predicted = self.predictor.predict_memory(deployment_name, self.namespace)
            else:
                predicted = None
            
            if predicted:
                # Calculate utilization ratio
                utilization = predicted / target if target > 0 else 0
                weighted_score += utilization * weight
                total_weight += weight
        
        if total_weight == 0:
            return current_replicas
        
        # Calculate scaling factor
        avg_utilization = weighted_score / total_weight
        scaling_factor = avg_utilization
        
        # Calculate desired replicas
        desired = int(current_replicas * scaling_factor)
        desired = max(min_replicas, min(desired, max_replicas))
        
        logger.info(f"{deployment_name}: current={current_replicas}, "
                   f"utilization={avg_utilization:.2f}, desired={desired}")
        
        return desired
    
    def run(self):
        """Main scaling loop"""
        logger.info("AI Scaler started")
        
        config = self.load_config()
        policies = config.get('scaling_policies', {})
        
        for deployment_name in policies.keys():
            try:
                desired = self.calculate_desired_replicas(deployment_name, config)
                current = self.get_deployment_replicas(deployment_name)
                
                if desired and desired != current:
                    self.scale_deployment(deployment_name, desired)
                    time.sleep(5)  # Small delay between scales
            except Exception as e:
                logger.error(f"Error processing {deployment_name}: {e}")

def main():
    namespace = os.getenv('NAMESPACE', 'kubernet-prod')
    config_map = os.getenv('CONFIG_MAP', 'ai-scaling-config')
    
    scaler = AIScaler(namespace, config_map)
    scaler.run()

if __name__ == '__main__':
    main()
