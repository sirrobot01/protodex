import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card.tsx';
import { Button } from '@/components/ui/button.tsx';
import { Input } from '@/components/ui/input.tsx';
import { Label } from '@/components/ui/label.tsx';
import { Textarea } from '@/components/ui/textarea.tsx';
import { Badge } from '@/components/ui/badge.tsx';
import { 
  Upload, 
  X, 
  FileText, 
  Package,
  Loader2
} from 'lucide-react';
import { useAuth } from '@/contexts/auth-context.tsx';
import { useToast } from '@/hooks/use-toast.ts';

interface PushFormData {
  packageName: string;
  version: string;
  description: string;
  tags: string[];
  files: File[];
}

export default function PushPackagePage() {
  const [formData, setFormData] = useState<PushFormData>({
    packageName: '',
    version: '',
    description: '',
    tags: [],
    files: []
  });
  const [tagInput, setTagInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [dragOver, setDragOver] = useState(false);
  const { authenticatedFetch } = useAuth();
  const { toast } = useToast();
  const navigate = useNavigate();

  const handleInputChange = (field: keyof PushFormData, value: string) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleAddTag = () => {
    if (tagInput.trim() && !formData.tags.includes(tagInput.trim())) {
      setFormData(prev => ({
        ...prev,
        tags: [...prev.tags, tagInput.trim()]
      }));
      setTagInput('');
    }
  };

  const handleRemoveTag = (tagToRemove: string) => {
    setFormData(prev => ({
      ...prev,
      tags: prev.tags.filter(tag => tag !== tagToRemove)
    }));
  };

  const handleFileSelection = (files: FileList) => {
    const projectFiles = Array.from(files).filter(file => 
      file.name.endsWith('.proto') || 
      file.name.endsWith('.yaml') || 
      file.name.endsWith('.yml') ||
      file.name === 'README.md'
    );
    
    setFormData(prev => ({
      ...prev,
      files: [...prev.files, ...projectFiles]
    }));
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    if (e.dataTransfer.files) {
      handleFileSelection(e.dataTransfer.files);
    }
  };

  const handleRemoveFile = (fileToRemove: File) => {
    setFormData(prev => ({
      ...prev,
      files: prev.files.filter(file => file !== fileToRemove)
    }));
  };

  const validateForm = (): string | null => {
    if (!formData.packageName.trim()) return 'Package name is required';
    if (!formData.version.trim()) return 'Version is required';
    if (formData.files.length === 0) return 'At least one file is required';
    
    // Check if package name is valid
    const nameRegex = /^[a-z0-9-_\/]+$/;
    if (!nameRegex.test(formData.packageName)) {
      return 'Package name can only contain lowercase letters, numbers, hyphens, underscores, and forward slashes';
    }

    // Check if version is valid semver
    const versionRegex = /^v?\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?$/;
    if (!versionRegex.test(formData.version)) {
      return 'Version must be in semantic version format (e.g., v1.0.0 or 1.0.0)';
    }

    return null;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const validationError = validateForm();
    if (validationError) {
      toast({
        title: "Validation Error",
        description: validationError,
        variant: "destructive",
      });
      return;
    }

    setLoading(true);

    try {
      // First, create or ensure package exists
      try {
        await authenticatedFetch(`/api/packages`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            name: formData.packageName,
            description: formData.description,
            tags: formData.tags
          }),
        });
      } catch (err) {
        // Package might already exist, continue
      }

      // Create zip file from selected files
      const zipBlob = await createProjectZip(formData.files);
      
      // Then push the version with zip
      const formDataToSend = new FormData();
      formDataToSend.append('version', formData.version);
      formDataToSend.append('zip', zipBlob, `${formData.packageName}-${formData.version}.zip`);

      const response = await authenticatedFetch(`/api/packages/${formData.packageName}/versions/`, {
        method: 'POST',
        body: formDataToSend,
      });

      if (response.ok) {
        toast({
          title: "Success",
          description: `Package ${formData.packageName}:${formData.version} pushed successfully!`,
        });
        navigate(`/package/${formData.packageName}`);
      } else {
        const error = await response.json();
        throw new Error(error.error || 'Push failed');
      }
    } catch (error) {
      console.error('Push failed:', error);
      toast({
        title: "Error",
        description: error instanceof Error ? error.message : "Failed to push package",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  // Create a zip file from the selected files
  const createProjectZip = async (files: File[]): Promise<Blob> => {
    // Import JSZip dynamically to reduce bundle size
    const JSZip = (await import('jszip')).default;
    const zip = new JSZip();

    // Add files to zip maintaining their relative paths/names
    for (const file of files) {
      zip.file(file.name, file);
    }

    // Generate zip file
    return zip.generateAsync({ type: 'blob' });
  };

  return (
    <div className="max-w-4xl mx-auto space-y-8">
      <div className="space-y-2">
        <h1 className="text-3xl font-bold">Push Package</h1>
        <p className="text-muted-foreground">
          Upload your protobuf files and configuration to the registry
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center">
              <Package className="h-5 w-5 mr-2" />
              Package Information
            </CardTitle>
            <CardDescription>
              Basic information about your package
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="packageName">Package Name *</Label>
                <Input
                  id="packageName"
                  value={formData.packageName}
                  onChange={(e) => handleInputChange('packageName', e.target.value)}
                  placeholder="e.g., user-service-api"
                  required
                />
                <p className="text-xs text-muted-foreground">
                  Use lowercase letters, numbers, hyphens, and forward slashes
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="version">Version *</Label>
                <Input
                  id="version"
                  value={formData.version}
                  onChange={(e) => handleInputChange('version', e.target.value)}
                  placeholder="e.g., v1.0.0"
                  required
                />
                <p className="text-xs text-muted-foreground">
                  Semantic version format (e.g., v1.0.0)
                </p>
              </div>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                value={formData.description}
                onChange={(e) => handleInputChange('description', e.target.value)}
                placeholder="Brief description of your package..."
                rows={3}
              />
            </div>

            <div className="space-y-2">
              <Label>Tags</Label>
              <div className="flex flex-wrap gap-2 mb-2">
                {formData.tags.map((tag) => (
                  <Badge key={tag} variant="secondary" className="flex items-center gap-1">
                    {tag}
                    <button
                      type="button"
                      onClick={() => handleRemoveTag(tag)}
                      className="ml-1 hover:text-destructive"
                    >
                      <X className="h-3 w-3" />
                    </button>
                  </Badge>
                ))}
              </div>
              <div className="flex gap-2">
                <Input
                  value={tagInput}
                  onChange={(e) => setTagInput(e.target.value)}
                  placeholder="Add a tag..."
                  onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), handleAddTag())}
                />
                <Button type="button" onClick={handleAddTag} variant="outline">
                  Add
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center">
              <Upload className="h-5 w-5 mr-2" />
              Files
            </CardTitle>
            <CardDescription>
              Upload your project files - they will be bundled maintaining directory structure
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div
              className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
                dragOver ? 'border-primary bg-primary/5' : 'border-muted-foreground/25'
              }`}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
            >
              <Upload className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
              <div className="space-y-2">
                <p className="text-lg font-medium">Drop files here or click to browse</p>
                <p className="text-muted-foreground">
                  Supports .proto, .yaml, .yml files and README.md
                </p>
              </div>
              <input
                type="file"
                multiple
                accept=".proto,.yaml,.yml,.md"
                onChange={(e) => e.target.files && handleFileSelection(e.target.files)}
                className="hidden"
                id="file-input"
              />
              <label htmlFor="file-input">
                <Button type="button" className="mt-4" asChild>
                  <span>Browse Files</span>
                </Button>
              </label>
            </div>

            {formData.files.length > 0 && (
              <div className="space-y-2">
                <h4 className="font-medium">Selected Files ({formData.files.length})</h4>
                <div className="space-y-2">
                  {formData.files.map((file, index) => (
                    <div
                      key={index}
                      className="flex items-center justify-between p-3 border rounded-lg"
                    >
                      <div className="flex items-center space-x-3">
                        <FileText className="h-4 w-4 text-muted-foreground" />
                        <div>
                          <p className="font-medium">{file.name}</p>
                          <p className="text-sm text-muted-foreground">
                            {formatFileSize(file.size)}
                          </p>
                        </div>
                      </div>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => handleRemoveFile(file)}
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        <div className="flex items-center justify-between">
          <Button
            type="button"
            variant="outline"
            onClick={() => navigate('/dashboard')}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={loading}>
            {loading ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Pushing...
              </>
            ) : (
              <>
                <Upload className="h-4 w-4 mr-2" />
                Push Package
              </>
            )}
          </Button>
        </div>
      </form>
    </div>
  );
}