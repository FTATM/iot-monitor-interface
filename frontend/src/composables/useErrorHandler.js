import { useI18n } from 'vue-i18n';
import { toast } from 'vue3-toastify';

export function useErrorHandler() {
  const { t } = useI18n();

  /**
   * Handles API errors, translates them, and triggers a toast notification.
   * 
   * @param {Object|String} error - The error object ref or string (e.g., deviceAddedError)
   * @param {String} fallbackKey - The i18n key for the default error (e.g., 'device.messages.exportFailed')
   * @param {Object} params - Dynamic parameters for the translation (e.g., { format: 'CSV', item: 'Sensor' })
   */
  const handleError = (error, fallbackKey, params = {}) => {
    // Extract the message from a ref object or use the string directly
    const backendMsg = error?.value?.message || error?.message || error;

    // Pass the params object directly to translate keys like "Failed to export {format} file"
    let errorMessage = t(fallbackKey, params);

    const displayItem = params.item || "";
    if (backendMsg === 't_dup') {
      errorMessage = t('common.messages.duplicateError', { item: displayItem });
    } else if (backendMsg === 't_no_access') {
      errorMessage = t('common.messages.noAccessError');
    } else if (backendMsg === 't_delete_in_used') {
      const displayItem = params.item || "";
      errorMessage = t('common.messages.deleteFailedInUsed', { item: displayItem });
    } else if (backendMsg) {
      errorMessage = backendMsg;
    }

    return errorMessage;
  };

  return { handleError };
}