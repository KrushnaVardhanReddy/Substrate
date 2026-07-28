const fs = require('fs');
const path = require('path');
const { PGlite } = require('@electric-sql/pglite');
const { createServer } = require('pglite-server');

async function main() {
  console.log("Starting PGlite in-memory database...");
  const db = new PGlite();
  
  // 1. Run migrations
  const migrationsDir = path.join(__dirname, '../../api/migrations');
  const schemaDir = path.join(__dirname, '../../api/internal/db/schema');

  const migrationFiles = fs.readdirSync(migrationsDir)
    .filter(f => f.endsWith('.up.sql'))
    .map(f => ({ path: path.join(migrationsDir, f), name: f }));
    
  const schemaFiles = fs.readdirSync(schemaDir)
    .filter(f => f.endsWith('.sql') && !f.endsWith('.down.sql'))
    .map(f => ({ path: path.join(schemaDir, f), name: f }));

  const files = [...migrationFiles, ...schemaFiles]
    .sort((a, b) => a.name.localeCompare(b.name)); // Sorts alphabetically which matches 0001, 0002 etc.

  // Deduplicate schemas that have already been applied by migrations
  const uniqueFilesMap = new Map();
  for (const file of files) {
    let sql = fs.readFileSync(file.path, 'utf8');
    // Polyfill uuid-ossp for PGlite (standard gen_random_uuid instead)
    sql = sql.replace(/CREATE EXTENSION IF NOT EXISTS "uuid-ossp";/g, '');
    sql = sql.replace(/uuid_generate_v4\(\)/g, 'gen_random_uuid()');
    // Basic hash to avoid duplicate execution
    const hash = require('crypto').createHash('sha256').update(sql).digest('hex');
    if (!uniqueFilesMap.has(hash)) {
      uniqueFilesMap.set(hash, { file, sql });
    }
  }

  const uniqueFiles = Array.from(uniqueFilesMap.values());
  console.log(`Found ${uniqueFiles.length} unique migration files. Applying...`);

  for (const { file, sql } of uniqueFiles) {
    try {
      await db.exec(sql);
    } catch (e) {
      console.error(`Failed to execute migration ${file.name}:`, e);
      process.exit(1);
    }
  }
  console.log("Migrations applied successfully.");

  // 2. Run seed data
  const seedFile = path.join(__dirname, '../../seed_mcp.sql');
  if (fs.existsSync(seedFile)) {
    console.log("Applying seed_mcp.sql...");
    const seedSql = fs.readFileSync(seedFile, 'utf8');
    try {
      await db.exec(seedSql);
      console.log("Seed data applied successfully.");
    } catch (e) {
      console.error(`Failed to execute seed data:`, e);
      process.exit(1);
    }
  } else {
    console.log("No seed_mcp.sql found, skipping.");
  }

  // 3. Start TCP server
  const server = createServer(db);
  server.listen(54320, () => {
    console.log("PGLITE_READY");
  });
  
  // Handle shutdown gracefully
  const shutdown = () => {
    server.close(() => {
      process.exit(0);
    });
  };
  
  process.on('SIGINT', shutdown);
  process.on('SIGTERM', shutdown);
}

main().catch(console.error);
