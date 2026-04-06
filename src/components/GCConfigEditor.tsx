import React, { ChangeEvent, useEffect, useState } from "react";
import { Alert, Legend, LegacyForms } from "@grafana/ui";
import { DataSourcePluginOptionsEditorProps } from "@grafana/data";
import { GCDataSourceOptions, GCJsonData, GCSecureJsonData } from "../types";
import {
  getAuthorizationValue,
  getHostnameValue,
  isValidApiHostname,
} from "../token";

const { FormField, SecretFormField } = LegacyForms;

const API_URL_VALIDATION_ERROR =
  "API URL must be a hostname only: start with api., end with .com, and include no path (for example api.gcore.com).";

interface Props
  extends DataSourcePluginOptionsEditorProps<GCDataSourceOptions> {}

export const GCConfigEditor = ({ options, onOptionsChange }: Props) => {
  const secureJsonData = (options.secureJsonData || {}) as GCSecureJsonData;
  const jsonData = (options.jsonData || {}) as GCJsonData;
  const [apiKey, setApiKey] = useState<string>(secureJsonData.apiKey || "");
  const [apiUrl, setApiUrl] = useState<string>(jsonData.apiUrl || "");
  const [apiUrlError, setApiUrlError] = useState<string | undefined>();

  useEffect(() => {
    const stored = jsonData.apiUrl || "";
    setApiUrl(stored);
    const normalized = getHostnameValue(stored.trim());
    if (normalized !== "" && !isValidApiHostname(normalized)) {
      setApiUrlError(API_URL_VALIDATION_ERROR);
    } else {
      setApiUrlError(undefined);
    }
  }, [jsonData.apiUrl]);

  const onApiUrlChange = (event: ChangeEvent<HTMLInputElement>) => {
    const next = event.target.value;
    setApiUrl(next);
    const normalized = getHostnameValue(next.trim());
    if (normalized === "" || isValidApiHostname(normalized)) {
      setApiUrlError(undefined);
    }
  };

  const updateApiUrl = () => {
    const normalizedApiUrl = getHostnameValue(apiUrl.trim());
    if (normalizedApiUrl !== "" && !isValidApiHostname(normalizedApiUrl)) {
      setApiUrlError(API_URL_VALIDATION_ERROR);
      return;
    }
    setApiUrlError(undefined);
    onOptionsChange({
      ...options,
      jsonData: { ...jsonData, apiUrl: normalizedApiUrl },
    });
  };

  const onApiKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    setApiKey(event.target.value);
  };

  const updateApiKey = () => {
    const normalizedApiKey = getAuthorizationValue(apiKey.trim());
    onOptionsChange({
      ...options,
      secureJsonData: { apiKey: normalizedApiKey },
    });
  };

  const onResetApiKey = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        apiKey: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        apiKey: "",
      },
    });
  };

  const { secureJsonFields } = options;
  const isConfigured = secureJsonFields && secureJsonFields.apiKey;

  return (
    <>
      <Legend>HTTP</Legend>

      <div className="gf-form-group">
        <FormField
          label={"URL"}
          labelWidth={8}
          inputWidth={20}
          placeholder={"API base url"}
          value={apiUrl}
          onChange={onApiUrlChange}
          onBlur={updateApiUrl}
          required={true}
        />
      </div>

      {apiUrlError && (
        <div className="gf-form-group">
          <Alert severity="error" title="Invalid API URL">
            {apiUrlError}
          </Alert>
        </div>
      )}

      <div className="gf-form-group">
        <SecretFormField
          isConfigured={isConfigured}
          label="API key"
          placeholder="Secure field"
          labelWidth={8}
          inputWidth={20}
          value={apiKey}
          onChange={onApiKeyChange}
          onBlur={updateApiKey}
          onReset={onResetApiKey}
        />
      </div>

      <div className="gf-form-group">
        <Alert severity={"info"} title="How to create a API token?">
          <a
            href="https://gcore.com/docs/account-settings/create-use-or-delete-a-permanent-api-token"
            target="_blank"
            rel="noreferrer"
          >
            https://gcore.com/docs/account-settings/create-use-or-delete-a-permanent-api-token
          </a>
        </Alert>
      </div>
    </>
  );
};
