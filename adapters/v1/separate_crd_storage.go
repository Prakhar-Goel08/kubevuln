package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kubescape/go-logger"
	"github.com/kubescape/go-logger/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// SeparateCRDStorageAdapter handles saving dive and trufflehog reports to separate Kubernetes CRDs
type SeparateCRDStorageAdapter struct {
	namespace         string
	client            dynamic.Interface
	secretScanningGVR schema.GroupVersionResource
	diveScanningGVR   schema.GroupVersionResource
}

// NewSeparateCRDStorageAdapter creates a new storage adapter for separate CRDs
func NewSeparateCRDStorageAdapter(namespace string) *SeparateCRDStorageAdapter {
	adapter := &SeparateCRDStorageAdapter{
		namespace: namespace,
		secretScanningGVR: schema.GroupVersionResource{
			Group:    "kubevuln.io",
			Version:  "v1",
			Resource: "secretscannings",
		},
		diveScanningGVR: schema.GroupVersionResource{
			Group:    "kubevuln.io",
			Version:  "v1",
			Resource: "divescannings",
		},
	}

	if err := adapter.initK8sClient(); err != nil {
		logger.L().Error("failed to initialize k8s client for separate CRD storage", helpers.Error(err))
		return adapter
	}

	logger.L().Info("separate CRD storage enabled successfully")
	return adapter
}

// initK8sClient initializes the Kubernetes client
func (s *SeparateCRDStorageAdapter) initK8sClient() error {
	var config *rest.Config
	var err error

	// Try in-cluster config first
	if config, err = rest.InClusterConfig(); err != nil {
		// Fall back to kubeconfig
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		}

		if config, err = clientcmd.BuildConfigFromFlags("", kubeconfig); err != nil {
			return fmt.Errorf("failed to get k8s config: %w", err)
		}
	}

	if s.client, err = dynamic.NewForConfig(config); err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return nil
}

// SaveDiveReport saves dive report to DiveScanning CRD
func (s *SeparateCRDStorageAdapter) SaveDiveReport(ctx context.Context, imageTag, imageName, jobID string, diveResult *DiveResult, outputPath string) error {
	if s.client == nil {
		return fmt.Errorf("k8s client not initialized")
	}

	// Try to save to CRD first
	if err := s.saveDiveReportToCRD(ctx, imageTag, imageName, jobID, diveResult, outputPath); err != nil {
		logger.L().Error("failed to save dive report to CRD, falling back to file storage", helpers.Error(err))
		// Fall back to file storage
		return s.saveDiveReportToFile(diveResult, outputPath)
	}

	return nil
}

// SaveSecretReport saves trufflehog report to SecretScanning CRD
func (s *SeparateCRDStorageAdapter) SaveSecretReport(ctx context.Context, imageTag, imageName, jobID string, truffleHogResults []TruffleHogResult, outputPath string) error {
	if s.client == nil {
		return fmt.Errorf("k8s client not initialized")
	}

	// Try to save to CRD first
	if err := s.saveSecretReportToCRD(ctx, imageTag, imageName, jobID, truffleHogResults, outputPath); err != nil {
		logger.L().Error("failed to save secret report to CRD, falling back to file storage", helpers.Error(err))
		// Fall back to file storage
		return s.saveSecretReportToFile(truffleHogResults, outputPath)
	}

	return nil
}

// saveDiveReportToCRD saves the dive report to DiveScanning CRD
func (s *SeparateCRDStorageAdapter) saveDiveReportToCRD(ctx context.Context, imageTag, imageName, jobID string, diveResult *DiveResult, outputPath string) error {
	sanitizedImageName := SanitizeLabel(imageName)
	sanitizedJobID := SanitizeLabel(jobID)

	// Read the complete dive report JSON file
	var diveReportJSON string
	if diveData, err := os.ReadFile(outputPath); err == nil {
		diveReportJSON = string(diveData)
		logger.L().Debug("read complete dive report JSON file", helpers.String("path", outputPath), helpers.Int("size", len(diveData)))
	} else {
		// If we can't read the file, marshal the result object
		if diveResult != nil {
			if data, err := json.Marshal(diveResult); err == nil {
				diveReportJSON = string(data)
			} else {
				return fmt.Errorf("failed to marshal dive result: %w", err)
			}
		} else {
			return fmt.Errorf("no dive data available")
		}
	}

	// Extract additional metadata from dive result
	var imageSize int64
	var layers int
	var efficiencyScore float64
	if diveResult != nil {
		imageSize = diveResult.Image.SizeBytes
		layers = len(diveResult.Layer)
		efficiencyScore = diveResult.Image.EfficiencyScore
	}

	// Create the dive scanning report object
	diveScanReport := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubevuln.io/v1",
			"kind":       "DiveScanning",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf("%s-%s-dive", sanitizedImageName, sanitizedJobID),
				"namespace": s.namespace,
				"labels": map[string]interface{}{
					"image":    sanitizedImageName,
					"jobId":    sanitizedJobID,
					"scanType": "dive",
					"tool":     "dive",
				},
				"annotations": map[string]interface{}{
					"kubevuln.io/dive-report": "Complete dive analysis JSON report",
					"kubevuln.io/scan-tool":   "dive",
				},
			},
			"spec": map[string]interface{}{
				"image":           imageTag,
				"namespace":       s.namespace,
				"timestamp":       metav1.Now().Format("2006-01-02T15:04:05Z"),
				"clusterName":     "default", // You can make this configurable
				"jobID":           sanitizedJobID,
				"diveReport":      diveReportJSON,
				"imageSize":       imageSize,
				"layers":          layers,
				"efficiencyScore": efficiencyScore,
				"reportPath":      outputPath,
			},
			"status": map[string]interface{}{
				"status":      "completed",
				"lastUpdated": metav1.Now().Format("2006-01-02T15:04:05Z"),
			},
		},
	}

	// Create the dive CRD resource
	_, err := s.client.Resource(s.diveScanningGVR).Namespace(s.namespace).Create(ctx, diveScanReport, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create dive scanning CRD: %w", err)
	}

	logger.L().Info("dive report saved to DiveScanning CRD successfully",
		helpers.String("image", imageTag),
		helpers.String("namespace", s.namespace),
		helpers.String("jobID", sanitizedJobID),
		helpers.Int("diveReportSize", len(diveReportJSON)))

	return nil
}

// saveSecretReportToCRD saves the trufflehog report to SecretScanning CRD
func (s *SeparateCRDStorageAdapter) saveSecretReportToCRD(ctx context.Context, imageTag, imageName, jobID string, truffleHogResults []TruffleHogResult, outputPath string) error {
	sanitizedImageName := SanitizeLabel(imageName)
	sanitizedJobID := SanitizeLabel(jobID)

	// Read the complete trufflehog report JSON file
	var truffleHogReportJSON string
	if truffleHogData, err := os.ReadFile(outputPath); err == nil {
		truffleHogReportJSON = string(truffleHogData)
		logger.L().Debug("read complete trufflehog report JSON file", helpers.String("path", outputPath), helpers.Int("size", len(truffleHogData)))
	} else {
		// If we can't read the file, marshal the results
		if truffleHogResults != nil {
			if data, err := json.Marshal(truffleHogResults); err == nil {
				truffleHogReportJSON = string(data)
			} else {
				return fmt.Errorf("failed to marshal trufflehog results: %w", err)
			}
		} else {
			// No secrets found
			truffleHogReportJSON = "[]"
		}
	}

	// Count secrets found
	secretsFound := len(truffleHogResults)

	// Create the secret scanning report object
	secretScanReport := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubevuln.io/v1",
			"kind":       "SecretScanning",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf("%s-%s-secret", sanitizedImageName, sanitizedJobID),
				"namespace": s.namespace,
				"labels": map[string]interface{}{
					"image":    sanitizedImageName,
					"jobId":    sanitizedJobID,
					"scanType": "secret",
					"tool":     "trufflehog",
				},
				"annotations": map[string]interface{}{
					"kubevuln.io/trufflehog-report": "Complete trufflehog secrets scan JSON report",
					"kubevuln.io/scan-tool":         "trufflehog",
				},
			},
			"spec": map[string]interface{}{
				"image":            imageTag,
				"namespace":        s.namespace,
				"timestamp":        metav1.Now().Format("2006-01-02T15:04:05Z"),
				"clusterName":      "default", // You can make this configurable
				"jobID":            sanitizedJobID,
				"trufflehogReport": truffleHogReportJSON,
				"secretsFound":     secretsFound,
				"reportPath":       outputPath,
			},
			"status": map[string]interface{}{
				"status":      "completed",
				"lastUpdated": metav1.Now().Format("2006-01-02T15:04:05Z"),
			},
		},
	}

	// Create the secret scanning CRD resource
	_, err := s.client.Resource(s.secretScanningGVR).Namespace(s.namespace).Create(ctx, secretScanReport, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create secret scanning CRD: %w", err)
	}

	logger.L().Info("trufflehog report saved to SecretScanning CRD successfully",
		helpers.String("image", imageTag),
		helpers.String("namespace", s.namespace),
		helpers.String("jobID", sanitizedJobID),
		helpers.Int("secretsFound", secretsFound),
		helpers.Int("trufflehogReportSize", len(truffleHogReportJSON)))

	return nil
}

// saveDiveReportToFile saves dive report to a JSON file as fallback
func (s *SeparateCRDStorageAdapter) saveDiveReportToFile(diveResult *DiveResult, outputPath string) error {
	data, err := json.MarshalIndent(diveResult, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal dive result: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write dive report to file: %w", err)
	}

	logger.L().Info("dive report saved to file", helpers.String("path", outputPath))
	return nil
}

// saveSecretReportToFile saves trufflehog report to a JSON file as fallback
func (s *SeparateCRDStorageAdapter) saveSecretReportToFile(truffleHogResults []TruffleHogResult, outputPath string) error {
	data, err := json.MarshalIndent(truffleHogResults, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal trufflehog results: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write trufflehog report to file: %w", err)
	}

	logger.L().Info("trufflehog report saved to file", helpers.String("path", outputPath))
	return nil
}
