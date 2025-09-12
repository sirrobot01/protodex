import {useEffect, useState} from 'react';
import {Link, useParams} from 'react-router-dom';
import {Card, CardContent, CardDescription, CardHeader, CardTitle} from '@/components/ui/card.tsx';
import {Button} from '@/components/ui/button.tsx';
import {Badge} from '@/components/ui/badge.tsx';
import {Tabs, TabsContent, TabsList, TabsTrigger} from '@/components/ui/tabs.tsx';
import {Separator} from '@/components/ui/separator.tsx';
import ReactMarkdown from 'react-markdown';
import Prism from 'prismjs';
import 'prismjs/themes/prism-tomorrow.css';
import 'prismjs/components/prism-protobuf';
import 'prismjs/components/prism-yaml';
import {
    Calendar,
    Code,
    Copy,
    Download,
    FileText,
    Folder,
    FolderOpen,
    History,
    Package,
    Settings,
    Terminal,
} from 'lucide-react';
import {useAuth} from '@/contexts/auth-context.tsx';
import {useToast} from '@/hooks/use-toast.ts';

interface PackageData {
  id: string;
  name: string;
  description: string;
  version: string;
  downloads: number;
  stars: number;
  tags: string[];
  created_at: string;
  updated_at: string;
  config?: {
    content: string;
    filename: string;
  };
  schema?: {
    files: Array<{
      name: string;
      content: string;
      path: string;
    }>;
  };
  versions?: Array<{
    version: string;
    created_at: string;
    downloads: number;
  }>;
}

export default function PackageDetailPage() {
  const { name } = useParams<{ name: string }>();
  const [pkg, setPkg] = useState<PackageData | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedFile, setSelectedFile] = useState<string | null>(null);
  const [expandedDirs, setExpandedDirs] = useState<Set<string>>(new Set());
  const [selectedVersion, setSelectedVersion] = useState<string | null>(null);
  const { authenticatedFetch } = useAuth();
  const { toast } = useToast();

  useEffect(() => {
    if (name) {
      fetchPackage(name);
    }
  }, [name]);

  useEffect(() => {
    // Set selectedVersion to current version when package is loaded
    if (pkg && !selectedVersion) {
      setSelectedVersion(pkg.version);
    }
  }, [pkg, selectedVersion]);

  useEffect(() => {
    if (pkg?.schema?.files && pkg.schema.files.length > 0 && !selectedFile) {
      // Filter out README and protodex.yaml files to match Schema tab display
      const filteredFiles = pkg.schema.files.filter(file => {
        const fileName = file.name.toLowerCase();
        const isReadme = fileName === 'readme.md' || fileName === 'readme' || fileName.startsWith('readme.');
        const isProtodexConfig = fileName === 'protodex.yaml' || fileName === 'protodex.yml';
        return !isReadme && !isProtodexConfig;
      });
      
      // Set first filtered file as selected and expand all directories in its path
      if (filteredFiles.length > 0) {
        const firstFile = filteredFiles[0];
        setSelectedFile(firstFile.path);
        
        // Auto-expand directories containing the first file
        const pathParts = firstFile.path.split('/');
        const newExpanded = new Set<string>();
        for (let i = 1; i < pathParts.length; i++) {
          newExpanded.add(pathParts.slice(0, i).join('/'));
        }
        setExpandedDirs(newExpanded);
      }
    }
  }, [pkg?.schema?.files, selectedFile]);

  const fetchPackage = async (packageName: string) => {
    try {
      // Fetch package info
      const packageResponse = await authenticatedFetch(`/api/packages/${packageName}`, {
        credentials: 'include',
      });
      
      if (!packageResponse.ok) {
        throw new Error('Package not found');
      }
      
      const packageData = await packageResponse.json();
      
      // Fetch versions
      const versionsResponse = await authenticatedFetch(`/api/packages/${packageName}/versions`, {
        credentials: 'include',
      });
      
      let versions = [];
      if (versionsResponse.ok) {
        versions = await versionsResponse.json() || [];
      }
      
      // Fetch schema for latest version if available
      let schemaData = null;
      if (versions.length > 0) {
        const latestVersion = versions[0]; // Assuming sorted by latest
        const schemaResponse = await authenticatedFetch(`/api/packages/${packageName}/versions/${latestVersion.version}/schema`, {
          credentials: 'include',
        });
        
        if (schemaResponse.ok) {
          schemaData = await schemaResponse.json();
        }
      }
      
      setPkg({
        ...packageData,
        version: versions.length > 0 ? versions[0].version : '1.0.0',
        versions: versions,
        schema: schemaData ? {
          files: schemaData.files.map((file: any) => ({
            name: file.name,
            content: file.content,
            path: file.path || file.name
          }))
        } : null,
        created_at: packageData.created_at,
        updated_at: packageData.created_at
      });
    } catch (error) {
      console.error('Failed to fetch package:', error);
      toast({
        title: "Error",
        description: "Failed to load package details",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast({
      title: "Copied",
      description: "Command copied to clipboard!",
    });
  };

  const downloadPackage = async () => {
    if (!pkg) return;
    
    try {
      const response = await authenticatedFetch(`/api/packages/${pkg.name}/versions/${pkg.version}/files`, {
        method: 'GET',
        credentials: 'include',
      });
      
      if (response.ok) {
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.style.display = 'none';
        a.href = url;
        a.download = `${pkg.name}-${pkg.version}.zip`;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
        
        toast({
          title: "Success",
          description: "Package downloaded successfully!",
        });
      } else {
        throw new Error('Download failed');
      }
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to download package",
        variant: "destructive",
      });
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  };

  // Build hierarchical file tree from flat file list
  const buildFileTree = (files: Array<{name: string; content: string; path: string}>) => {
    const tree: any = {};
    
    files.forEach(file => {
      const pathParts = file.path.split('/');
      let current = tree;
      
      // Build the tree structure
      pathParts.forEach((part, index) => {
        if (index === pathParts.length - 1) {
          // This is the file
          current[part] = {
            type: 'file',
            name: file.name,
            path: file.path,
            content: file.content
          };
        } else {
          // This is a directory
          if (!current[part]) {
            current[part] = {
              type: 'directory',
              children: {}
            };
          }
          current = current[part].children;
        }
      });
    });
    
    return tree;
  };

  // Flatten tree back to file list for easier rendering
  const flattenTree = (tree: any, prefix = ''): Array<{name: string; path: string; type: 'file' | 'directory'; indent: number}> => {
    const result: Array<{name: string; path: string; type: 'file' | 'directory'; indent: number}> = [];
    
    Object.keys(tree).sort().forEach(key => {
      const item = tree[key];
      const currentPath = prefix ? `${prefix}/${key}` : key;
      const indent = prefix.split('/').length;
      
      if (item.type === 'directory') {
        result.push({
          name: key,
          path: currentPath,
          type: 'directory',
          indent
        });
        result.push(...flattenTree(item.children, currentPath));
      } else {
        result.push({
          name: key,
          path: item.path,
          type: 'file',
          indent
        });
      }
    });
    
    return result;
  };

  // Toggle directory expansion
  const toggleDirectory = (path: string) => {
    const newExpanded = new Set(expandedDirs);
    if (newExpanded.has(path)) {
      newExpanded.delete(path);
    } else {
      newExpanded.add(path);
    }
    setExpandedDirs(newExpanded);
  };

  // Get hierarchical file list for display (excluding README and protodex.yaml)
  const getHierarchicalFiles = () => {
    if (!pkg?.schema?.files) return [];
    
    // Filter out README and protodex.yaml files
    const filteredFiles = pkg.schema.files.filter(file => {
      const fileName = file.name.toLowerCase();
      const isReadme = fileName === 'readme.md' || fileName === 'readme' || fileName.startsWith('readme.');
      const isProtodexConfig = fileName === 'protodex.yaml' || fileName === 'protodex.yml';
      return !isReadme && !isProtodexConfig;
    });
    
    const tree = buildFileTree(filteredFiles);
    const flatFiles = flattenTree(tree);
    
    // Filter out collapsed directories
    const visibleFiles: typeof flatFiles = [];
    
    flatFiles.forEach(file => {
      const pathParts = file.path.split('/');
      
      // Check if all parent directories are expanded
      let shouldShow = true;
      for (let i = 1; i <= pathParts.length - 1; i++) {
        const ancestorPath = pathParts.slice(0, i).join('/');
        if (!expandedDirs.has(ancestorPath)) {
          shouldShow = false;
          break;
        }
      }
      
      if (shouldShow) {
        visibleFiles.push(file);
      }
    });
    
    return visibleFiles;
  };

  // Load schema for a specific version
  const loadVersionSchema = async (packageName: string, version: string) => {
    try {
      const schemaResponse = await authenticatedFetch(`/api/packages/${packageName}/versions/${version}/schema`, {
        credentials: 'include',
      });
      
      if (schemaResponse.ok) {
        const schemaData = await schemaResponse.json();
        return {
          files: schemaData.files.map((file: any) => ({
            name: file.name,
            content: file.content,
            path: file.path || file.name
          }))
        };
      }
      return null;
    } catch (error) {
      console.error('Failed to fetch version schema:', error);
      return null;
    }
  };

  // Switch to a different version
  const switchToVersion = async (version: string) => {
    if (!name || version === selectedVersion) return;
    
    setSelectedVersion(version);
    setSelectedFile(null);
    setExpandedDirs(new Set());
    
    const schemaData = await loadVersionSchema(name, version);
    if (schemaData && pkg) {
      setPkg({
        ...pkg,
        version: version,
        schema: schemaData
      });
    }
  };

  // Helper function to find protodex.yaml file from schema files
  const findProtodexConfig = () => {
    if (!pkg?.schema?.files) return null;
    return pkg.schema.files.find(file => 
      file.name.toLowerCase() === 'protodex.yaml' || 
      file.name.toLowerCase() === 'protodex.yml'
    );
  };

  // Helper function to find README file from schema files
  const findReadmeFile = () => {
    if (!pkg?.schema?.files) return null;
    return pkg.schema.files.find(file => {
      const fileName = file.name.toLowerCase();
      return fileName === 'readme.md' || 
             fileName === 'readme' || 
             fileName.startsWith('readme.');
    });
  };

  // Helper function to get file language for syntax highlighting
  const getFileLanguage = (fileName: string) => {
    const extension = fileName.split('.').pop()?.toLowerCase();
    switch (extension) {
      case 'proto':
        return 'protobuf';
      case 'yaml':
      case 'yml':
        return 'yaml';
      case 'json':
        return 'json';
      case 'js':
      case 'jsx':
        return 'javascript';
      case 'ts':
      case 'tsx':
        return 'typescript';
      case 'go':
        return 'go';
      case 'py':
        return 'python';
      default:
        return 'text';
    }
  };

  // Syntax highlight code using Prism
  const highlightCode = (code: string, language: string) => {
    if (language === 'text') {
      return code;
    }
    
    try {
        return Prism.highlight(code, Prism.languages[language] || Prism.languages.text, language);
    } catch (error) {
      return code;
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    );
  }

  if (!pkg) {
    return (
      <div className="text-center py-12">
        <Package className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
        <h3 className="text-lg font-medium mb-2">Package not found</h3>
        <p className="text-muted-foreground">
          The package you're looking for doesn't exist or you don't have permission to view it.
        </p>
        <Link to="/dashboard" className="mt-4 inline-block">
          <Button>Back to Dashboard</Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="flex items-start justify-between">
        <div className="space-y-2">
          <h1 className="text-3xl font-bold">{pkg.name}</h1>
          <div className="flex items-center space-x-4 text-muted-foreground">
            <div className="flex items-center">
              <Calendar className="h-4 w-4 mr-1" />
              Created {formatDate(pkg.created_at)}
            </div>
            <div>{pkg.version}</div>
          </div>
          <p className="text-lg text-muted-foreground max-w-2xl">
            {pkg.description}
          </p>
          
          {/* Version Selector */}
          {pkg.versions && pkg.versions.length > 1 && (
            <div className="flex items-center space-x-2 mt-4">
              <History className="h-4 w-4 text-muted-foreground" />
              <span className="text-sm text-muted-foreground">Version:</span>
              <select
                value={selectedVersion || pkg.version}
                onChange={(e) => switchToVersion(e.target.value)}
                className="w-48 px-3 py-2 border border-input bg-background text-sm rounded-md focus:outline-none focus:ring-2 focus:ring-ring"
              >
                {pkg.versions.map((version) => (
                  <option key={version.version} value={version.version}>
                    {version.version} - {formatDate(version.created_at)}
                  </option>
                ))}
              </select>
            </div>
          )}
        </div>
        
        <div className="flex items-center space-x-2">
          <Button onClick={downloadPackage}>
            <Download className="h-4 w-4 mr-2" />
            Download
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Latest Version</CardTitle>
            <Package className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">v{pkg.version}</div>
            <p className="text-xs text-muted-foreground">
              Published {formatDate(pkg.created_at)}
            </p>
          </CardContent>
        </Card>
      </div>

      {pkg.tags.length > 0 && (
        <div className="space-y-2">
          <h3 className="text-lg font-semibold">Tags</h3>
          <div className="flex flex-wrap gap-2">
            {pkg.tags.map((tag) => (
              <Badge key={tag} variant="outline">
                {tag}
              </Badge>
            ))}
          </div>
        </div>
      )}

      <Tabs defaultValue="readme" className="space-y-4">
        <TabsList>
          <TabsTrigger value="readme">
            <FileText className="h-4 w-4 mr-2" />
            README
          </TabsTrigger>
          <TabsTrigger value="schema">
            <Code className="h-4 w-4 mr-2" />
            Schema
          </TabsTrigger>
          <TabsTrigger value="config">
            <Settings className="h-4 w-4 mr-2" />
            Config
          </TabsTrigger>
          <TabsTrigger value="install">
            <Terminal className="h-4 w-4 mr-2" />
            Installation
          </TabsTrigger>
          <TabsTrigger value="versions">
            <History className="h-4 w-4 mr-2" />
            Versions
          </TabsTrigger>
        </TabsList>

        <TabsContent value="readme">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>README</CardTitle>
                  <CardDescription>
                    Package documentation and usage information
                  </CardDescription>
                </div>
                {(() => {
                  const readmeFile = findReadmeFile();
                  return readmeFile ? (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => copyToClipboard(readmeFile.content)}
                    >
                      <Copy className="h-4 w-4" />
                    </Button>
                  ) : null;
                })()}
              </div>
            </CardHeader>
            <CardContent>
              {(() => {
                const readmeFile = findReadmeFile();
                
                if (readmeFile) {
                  // Check if README is markdown
                  const isMarkdown = readmeFile.name.toLowerCase().endsWith('.md');
                  
                  if (isMarkdown) {
                    // Render markdown content
                    return (
                      <div className="space-y-4">
                        <div className="text-sm text-muted-foreground">
                        </div>
                        <div className="prose prose-sm max-w-none dark:prose-invert">
                          <ReactMarkdown>{readmeFile.content}</ReactMarkdown>
                        </div>
                      </div>
                    );
                  } else {
                    // Show plain text content
                    return (
                      <div className="space-y-4">
                        <div className="text-sm text-muted-foreground">
                        </div>
                        <pre className="bg-slate-50 dark:bg-slate-800 p-4 rounded-lg overflow-x-auto whitespace-pre-wrap">
                          <code className="text-sm text-slate-100">{readmeFile.content}</code>
                        </pre>
                      </div>
                    );
                  }
                } else {
                  // Fall back to static content
                  return (
                    <div className="prose prose-sm max-w-none dark:prose-invert">
                      <p>
                        This is a protocol buffer package that contains protobuf definitions
                        for {pkg.name}. Use the installation instructions to add this package
                        to your project.
                      </p>
                      <h3>Description</h3>
                      <p>{pkg.description}</p>
                      <h3>Usage</h3>
                      <p>
                        After installing this package, you can use the protobuf definitions
                        in your code. Refer to the schema tab to view the available message types
                        and services.
                      </p>
                    </div>
                  );
                }
              })()}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="schema">
          <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
            <Card className="lg:col-span-1">
              <CardHeader>
                <CardTitle className="text-base">Proto Files</CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                {pkg.schema?.files && pkg.schema.files.length > 0 ? (
                  <div className="space-y-1">
                    {getHierarchicalFiles().map((item) => (
                      <div
                        key={item.path}
                        style={{ marginLeft: `${item.indent * 16}px` }}
                      >
                        {item.type === 'directory' ? (
                          <button
                            onClick={() => toggleDirectory(item.path)}
                            className="w-full text-left px-4 py-2 text-sm hover:bg-accent transition-colors flex items-center"
                          >
                            {expandedDirs.has(item.path) ? (
                              <FolderOpen className="h-4 w-4 mr-2 text-muted-foreground" />
                            ) : (
                              <Folder className="h-4 w-4 mr-2 text-muted-foreground" />
                            )}
                            {item.name}
                          </button>
                        ) : (
                          <button
                            onClick={() => setSelectedFile(item.path)}
                            className={`w-full text-left px-4 py-2 text-sm hover:bg-accent transition-colors ${
                              selectedFile === item.path ? 'bg-accent' : ''
                            }`}
                          >
                            <div className="flex items-center">
                              <FileText className="h-4 w-4 mr-2 text-muted-foreground" />
                              {item.name}
                            </div>
                          </button>
                        )}
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="p-4 text-sm text-muted-foreground text-center">
                    No schema files available
                  </div>
                )}
              </CardContent>
            </Card>

            <Card className="lg:col-span-3">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base">
                    {selectedFile || 'Select a file'}
                  </CardTitle>
                  {selectedFile && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        const file = pkg.schema?.files.find(f => f.path === selectedFile);
                        if (file) copyToClipboard(file.content);
                      }}
                    >
                      <Copy className="h-4 w-4" />
                    </Button>
                  )}
                </div>
              </CardHeader>
              <CardContent>
                {selectedFile && pkg.schema?.files ? (
                  <div className="relative">
                    {(() => {
                      const file = pkg.schema.files.find(f => f.path === selectedFile);
                      if (!file) return null;
                      
                      const language = getFileLanguage(file.name);
                      const highlightedCode = highlightCode(file.content, language);
                      
                      return (
                        <pre className="bg-slate-50 dark:bg-slate-800 p-4 rounded-lg overflow-x-auto">
                          <code 
                            className={`text-sm language-${language} text-slate-100`}
                            dangerouslySetInnerHTML={{ __html: highlightedCode }}
                          />
                        </pre>
                      );
                    })()} 
                  </div>
                ) : (
                  <div className="text-center py-12 text-muted-foreground">
                    Select a proto file to view its contents
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="config">
          <Card>
            <CardHeader>
              <CardTitle>Package Configuration</CardTitle>
              <CardDescription>
                protodex.yaml configuration file
              </CardDescription>
            </CardHeader>
            <CardContent>
              {(() => {
                const protodexFile = findProtodexConfig();
                const configContent = protodexFile || pkg.config;
                
                return configContent ? (
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <h4 className="font-medium">
                        {protodexFile ? protodexFile.name : pkg.config?.filename}
                      </h4>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => copyToClipboard(protodexFile ? protodexFile.content : pkg.config!.content)}
                      >
                        <Copy className="h-4 w-4" />
                      </Button>
                    </div>
                    {(() => {
                      const content = protodexFile ? protodexFile.content : pkg.config?.content;
                      const fileName = protodexFile ? protodexFile.name : pkg.config?.filename || 'protodex.yaml';
                      const language = getFileLanguage(fileName);
                      const highlightedCode = highlightCode(content || '', language);

                      return (
                        <pre className="bg-slate-50 dark:bg-slate-800 p-4 rounded-lg overflow-x-auto">
                          <code
                            className={`text-sm language-${language} text-slate-100`}
                            dangerouslySetInnerHTML={{ __html: highlightedCode }}
                          />
                        </pre>
                      );
                    })()}
                  </div>
                ) : (
                  <div className="text-center py-12 text-muted-foreground">
                    No configuration file available
                  </div>
                );
              })()}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="install">
          <Card>
            <CardHeader>
              <CardTitle>Installation</CardTitle>
              <CardDescription>
                How to use this package in your project
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-6">
                <div>
                  <h4 className="font-medium mb-2">Using Protodex CLI</h4>
                  <div className="bg-muted p-4 rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                      <code className="text-sm">protodex pull {pkg.name}:{pkg.version}</code>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => copyToClipboard(`protodex pull ${pkg.name}:${pkg.version}`)}
                      >
                        <Copy className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </div>

                <Separator />

                <div>
                  <h4 className="font-medium mb-2">Manual Download</h4>
                  <p className="text-sm text-muted-foreground mb-4">
                    Download the package as a ZIP file and extract it to your project
                  </p>
                  <Button onClick={downloadPackage}>
                    <Download className="h-4 w-4 mr-2" />
                    Download {pkg.name}-{pkg.version}.zip
                  </Button>
                </div>

                <Separator />

                <div>
                  <h4 className="font-medium mb-2">Registry Information</h4>
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Package:</span>
                      <span>{pkg.name}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Version:</span>
                      <span>{pkg.version}</span>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="versions">
          <Card>
            <CardHeader>
              <CardTitle>Version History</CardTitle>
              <CardDescription>
                All available versions of this package
              </CardDescription>
            </CardHeader>
            <CardContent>
              {pkg.versions && pkg.versions.length > 0 ? (
                <div className="space-y-4">
                  {pkg.versions.map((version, index) => (
                    <div
                      key={version.version}
                      className={`p-4 border rounded-lg transition-colors ${
                        selectedVersion === version.version || (!selectedVersion && index === 0)
                          ? 'border-primary bg-primary/5'
                          : 'border-border hover:border-accent-foreground/20'
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <div className="space-y-1">
                          <div className="flex items-center space-x-2">
                            <h4 className="font-medium">{version.version}</h4>
                            {index === 0 && (
                              <Badge variant="outline">Latest</Badge>
                            )}
                            {selectedVersion === version.version && (
                              <Badge>Current</Badge>
                            )}
                          </div>
                          <div className="flex items-center space-x-4 text-sm text-muted-foreground">
                            <div className="flex items-center">
                              <Calendar className="h-3 w-3 mr-1" />
                              {formatDate(version.created_at)}
                            </div>
                            <div className="flex items-center">
                              <Download className="h-3 w-3 mr-1" />
                              {version.downloads || 0} downloads
                            </div>
                          </div>
                        </div>
                        <div className="flex items-center space-x-2">
                          {selectedVersion !== version.version && (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => switchToVersion(version.version)}
                            >
                              View
                            </Button>
                          )}
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={async () => {
                              try {
                                const response = await authenticatedFetch(`/api/packages/${pkg.name}/versions/${version.version}/files`, {
                                  method: 'GET',
                                  credentials: 'include',
                                });
                                
                                if (response.ok) {
                                  const blob = await response.blob();
                                  const url = window.URL.createObjectURL(blob);
                                  const a = document.createElement('a');
                                  a.style.display = 'none';
                                  a.href = url;
                                  a.download = `${pkg.name}-${version.version}.zip`;
                                  document.body.appendChild(a);
                                  a.click();
                                  window.URL.revokeObjectURL(url);
                                  document.body.removeChild(a);
                                  
                                  toast({
                                    title: "Success",
                                    description: `Downloaded ${pkg.name} v${version.version}`,
                                  });
                                }
                              } catch (error) {
                                toast({
                                  title: "Error",
                                  description: "Failed to download version",
                                  variant: "destructive",
                                });
                              }
                            }}
                          >
                            <Download className="h-3 w-3" />
                          </Button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-12 text-muted-foreground">
                  <History className="h-12 w-12 mx-auto mb-4" />
                  <p>No version history available</p>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}