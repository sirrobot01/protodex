import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card.tsx';
import { Button } from '@/components/ui/button.tsx';
import { Input } from '@/components/ui/input.tsx';
import { Label } from '@/components/ui/label.tsx';
import { Textarea } from '@/components/ui/textarea.tsx';
import { Badge } from '@/components/ui/badge.tsx';
import { useToast } from '@/hooks/use-toast.ts';
import { Package, X, Plus } from 'lucide-react';
import {useAuth} from "@/contexts/auth-context.tsx";

interface CreatePackageFormData {
  name: string;
  description: string;
  version: string;
  tags: string;
}

export default function CreatePackagePage() {
  const [isLoading, setIsLoading] = useState(false);
  const [tags, setTags] = useState<string[]>([]);
  const [currentTag, setCurrentTag] = useState('');
  const { toast } = useToast();
  const navigate = useNavigate();
  const { authenticatedFetch } = useAuth();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<CreatePackageFormData>({
    defaultValues: {
      version: '1.0.0',
    }
  });

  const addTag = () => {
    if (currentTag.trim() && !tags.includes(currentTag.trim()) && tags.length < 10) {
      setTags([...tags, currentTag.trim()]);
      setCurrentTag('');
    }
  };

  const removeTag = (tagToRemove: string) => {
    setTags(tags.filter(tag => tag !== tagToRemove));
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault();
      addTag();
    }
  };

  const onSubmit = async (data: CreatePackageFormData) => {
    try {
      setIsLoading(true);
      
      const packageData = {
        ...data,
        tags: tags,
      };

      const response = await authenticatedFetch('/api/packages', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify(packageData),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to create package');
      }

      const result = await response.json();
      
      toast({
        title: "Success",
        description: "Package created successfully!",
      });
      
      navigate(`/package/${result.id}`);
    } catch (error) {
      toast({
        title: "Error",
        description: error instanceof Error ? error.message : 'Failed to create package',
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto space-y-8">
      <div className="text-center">
        <Package className="h-12 w-12 mx-auto mb-4 text-primary" />
        <h1 className="text-3xl font-bold">Create New Package</h1>
        <p className="text-muted-foreground mt-2">
          Create an empty package structure that you can populate via CLI
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Package Information</CardTitle>
          <CardDescription>
            Provide basic information about your protocol buffer package.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="name">Package Name</Label>
              <Input
                id="name"
                placeholder="my-awesome-package"
                {...register('name', {
                  required: 'Package name is required',
                  pattern: {
                    value: /^[a-z0-9-]+$/,
                    message: 'Package name can only contain lowercase letters, numbers, and hyphens'
                  },
                  minLength: {
                    value: 3,
                    message: 'Package name must be at least 3 characters'
                  }
                })}
              />
              {errors.name && (
                <p className="text-sm text-destructive">{errors.name.message}</p>
              )}
              <p className="text-xs text-muted-foreground">
                Use lowercase letters, numbers, and hyphens only
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                placeholder="A brief description of what this package contains..."
                className="min-h-[100px]"
                {...register('description', {
                  required: 'Description is required',
                  minLength: {
                    value: 10,
                    message: 'Description must be at least 10 characters'
                  }
                })}
              />
              {errors.description && (
                <p className="text-sm text-destructive">{errors.description.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="version">Initial Version</Label>
              <Input
                id="version"
                placeholder="1.0.0"
                {...register('version', {
                  required: 'Version is required',
                  pattern: {
                    value: /^\d+\.\d+\.\d+$/,
                    message: 'Version must follow semantic versioning (e.g., 1.0.0)'
                  }
                })}
              />
              {errors.version && (
                <p className="text-sm text-destructive">{errors.version.message}</p>
              )}
              <p className="text-xs text-muted-foreground">
                Use semantic versioning (major.minor.patch)
              </p>
            </div>

            <div className="space-y-2">
              <Label>Tags</Label>
              <div className="space-y-2">
                <div className="flex space-x-2">
                  <Input
                    placeholder="Add a tag..."
                    value={currentTag}
                    onChange={(e) => setCurrentTag(e.target.value)}
                    onKeyDown={handleKeyDown}
                    disabled={tags.length >= 10}
                  />
                  <Button
                    type="button"
                    variant="outline"
                    onClick={addTag}
                    disabled={!currentTag.trim() || tags.includes(currentTag.trim()) || tags.length >= 10}
                  >
                    <Plus className="h-4 w-4" />
                  </Button>
                </div>
                
                {tags.length > 0 && (
                  <div className="flex flex-wrap gap-2">
                    {tags.map((tag) => (
                      <Badge key={tag} variant="secondary" className="text-sm">
                        {tag}
                        <button
                          type="button"
                          onClick={() => removeTag(tag)}
                          className="ml-2 hover:text-destructive"
                        >
                          <X className="h-3 w-3" />
                        </button>
                      </Badge>
                    ))}
                  </div>
                )}
                
                <p className="text-xs text-muted-foreground">
                  Add up to 10 tags to help others discover your package. Press Enter or comma to add.
                </p>
              </div>
            </div>


            <div className="flex space-x-4">
              <Button type="submit" disabled={isLoading} className="flex-1">
                {isLoading ? 'Creating Package...' : 'Create Package'}
              </Button>
              <Button 
                type="button" 
                variant="outline" 
                onClick={() => navigate('/dashboard')}
              >
                Cancel
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Next Steps</CardTitle>
          <CardDescription>
            After creating your package, you can populate it using the Protodex CLI
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="rounded-lg border p-4">
              <h4 className="font-medium mb-2">1. Install Protodex CLI</h4>
              <pre className="text-sm bg-muted p-2 rounded overflow-x-auto">
                <code>go install github.com/sirrobot01/protodex@latest</code>
              </pre>
            </div>

            <div className="rounded-lg border p-4">
              <h4 className="font-medium mb-2">2. Initialize your package locally</h4>
              <pre className="text-sm bg-muted p-2 rounded overflow-x-auto">
                <code>protodex init your-package-name</code>
              </pre>
            </div>

            <div className="rounded-lg border p-4">
              <h4 className="font-medium mb-2">3. Add your .proto files and push</h4>
              <pre className="text-sm bg-muted p-2 rounded overflow-x-auto">
                <code>protodex push v1.0.0</code>
              </pre>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}