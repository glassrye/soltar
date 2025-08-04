#!/usr/bin/env node

const https = require('https');

class InfrastructureManager {
  constructor(baseURL, token) {
    this.baseURL = baseURL;
    this.token = token;
  }

  async makeRequest(path, method = 'GET', data = null) {
    return new Promise((resolve, reject) => {
      const options = {
        hostname: new URL(this.baseURL).hostname,
        port: 443,
        path: path,
        method: method,
        headers: {
          'Authorization': `Bearer ${this.token}`,
          'Content-Type': 'application/json'
        }
      };

      const req = https.request(options, (res) => {
        let body = '';
        res.on('data', (chunk) => {
          body += chunk;
        });
        res.on('end', () => {
          try {
            const response = JSON.parse(body);
            resolve({ status: res.statusCode, data: response });
          } catch (error) {
            resolve({ status: res.statusCode, data: body });
          }
        });
      });

      req.on('error', (error) => {
        reject(error);
      });

      if (data) {
        req.write(JSON.stringify(data));
      }

      req.end();
    });
  }

  async getInfrastructure() {
    console.log('📊 Getting infrastructure...');
    const response = await this.makeRequest('/infrastructure');
    
    if (response.status === 200) {
      console.log('✅ Infrastructure retrieved:');
      console.log(JSON.stringify(response.data, null, 2));
      return response.data;
    } else {
      console.error('❌ Failed to get infrastructure:', response.data);
      return null;
    }
  }

  async updateInfrastructure(infrastructure) {
    console.log('🔄 Updating infrastructure...');
    const response = await this.makeRequest('/infrastructure', 'POST', {
      infrastructure
    });
    
    if (response.status === 200) {
      console.log('✅ Infrastructure updated successfully');
      return response.data;
    } else {
      console.error('❌ Failed to update infrastructure:', response.data);
      return null;
    }
  }

  async addVPNInstance(instanceId) {
    const current = await this.getInfrastructure();
    if (!current) return;

    const updatedInfrastructure = {
      ...current.infrastructure,
      vpn_instances: [...current.infrastructure.vpn_instances, instanceId]
    };

    return await this.updateInfrastructure(updatedInfrastructure);
  }

  async addLoadBalancer(lbId) {
    const current = await this.getInfrastructure();
    if (!current) return;

    const updatedInfrastructure = {
      ...current.infrastructure,
      load_balancers: [...current.infrastructure.load_balancers, lbId]
    };

    return await this.updateInfrastructure(updatedInfrastructure);
  }

  async addDatabase(dbId) {
    const current = await this.getInfrastructure();
    if (!current) return;

    const updatedInfrastructure = {
      ...current.infrastructure,
      databases: [...current.infrastructure.databases, dbId]
    };

    return await this.updateInfrastructure(updatedInfrastructure);
  }

  async addStorage(storageId) {
    const current = await this.getInfrastructure();
    if (!current) return;

    const updatedInfrastructure = {
      ...current.infrastructure,
      storage: [...current.infrastructure.storage, storageId]
    };

    return await this.updateInfrastructure(updatedInfrastructure);
  }
}

// CLI Interface
async function main() {
  const args = process.argv.slice(2);
  
  if (args.length < 2) {
    console.log('Usage: node manage-infrastructure.js <base-url> <token> [command] [resource-id]');
    console.log('');
    console.log('Commands:');
    console.log('  get                    - Get current infrastructure');
    console.log('  add-vpn <instance-id>  - Add VPN instance');
    console.log('  add-lb <lb-id>         - Add load balancer');
    console.log('  add-db <db-id>         - Add database');
    console.log('  add-storage <storage-id> - Add storage');
    console.log('');
    console.log('Examples:');
    console.log('  node manage-infrastructure.js https://your-worker.workers.dev your-jwt-token get');
    console.log('  node manage-infrastructure.js https://your-worker.workers.dev your-jwt-token add-vpn vpn-instance-1');
    process.exit(1);
  }

  const [baseURL, token, command, resourceId] = args;
  const manager = new InfrastructureManager(baseURL, token);

  try {
    switch (command) {
      case 'get':
        await manager.getInfrastructure();
        break;
      case 'add-vpn':
        if (!resourceId) {
          console.error('❌ VPN instance ID required');
          process.exit(1);
        }
        await manager.addVPNInstance(resourceId);
        break;
      case 'add-lb':
        if (!resourceId) {
          console.error('❌ Load balancer ID required');
          process.exit(1);
        }
        await manager.addLoadBalancer(resourceId);
        break;
      case 'add-db':
        if (!resourceId) {
          console.error('❌ Database ID required');
          process.exit(1);
        }
        await manager.addDatabase(resourceId);
        break;
      case 'add-storage':
        if (!resourceId) {
          console.error('❌ Storage ID required');
          process.exit(1);
        }
        await manager.addStorage(resourceId);
        break;
      default:
        console.error('❌ Unknown command:', command);
        process.exit(1);
    }
  } catch (error) {
    console.error('❌ Error:', error.message);
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}

module.exports = InfrastructureManager; 